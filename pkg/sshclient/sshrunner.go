package sshclient

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sshpky/pkg/config"
	"sshpky/pkg/km"
	"sshpky/pkg/logger"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

// establishSSHClient sets up and dials an SSH connection, handling all auth.
func EstablishSSHClient(hostAlias string) (*ssh.Client, config.SshConfigItem, error) {
	return establishSSHClient(hostAlias, nil)
}

func establishSSHClient(hostAlias string, fallbackConn *config.SshConfigItem) (*ssh.Client, config.SshConfigItem, error) {
	ms := config.NewSSHConfigManager()
	cnf, err := ms.FindConfig(hostAlias)
	var conn config.SshConfigItem
	if err == nil && cnf != nil {
		conn = *cnf
	} else if fallbackConn != nil {
		conn = *fallbackConn
		if conn.Host == "" {
			conn.Host = hostAlias
		}
		if conn.HostName == "" {
			conn.HostName = conn.Host
		}
		if conn.Port == 0 {
			conn.Port = 22
		}
		logger.Debug("ssh: config for host alias %s not found, using runtime connection parameters", hostAlias)
	} else {
		return nil, config.SshConfigItem{}, fmt.Errorf("failed to find config for host alias '%s': %w", hostAlias, err)
	}

	if conn.User == "" {
		currentUser, err := user.Current()
		if err == nil {
			logger.Debug("ssh: user not specified, defaulting to current user %s", currentUser.Username)
			conn.User = currentUser.Username
		} else {
			logger.Debug("ssh: warning: could not get current user: %v", err)
		}
	}

	var usedPassword, usedMfaSecret string
	authMethods := []ssh.AuthMethod{}

	if conn.IdentityFile != "" {
		identityFile := conn.IdentityFile
		if strings.HasPrefix(identityFile, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, conn, fmt.Errorf("failed to get user home directory to expand path: %w", err)
			}
			identityFile = filepath.Join(home, identityFile[2:])
		}
		logger.Debug("auth: IdentityFile specified (%s), using public key authentication exclusively.", conn.IdentityFile)
		key, err := os.ReadFile(identityFile)
		if err != nil {
			return nil, conn, fmt.Errorf("failed to read identity file %s: %w", identityFile, err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, conn, fmt.Errorf("failed to parse private key from %s: %w. Key may be passphrase-protected, not yet supported", conn.IdentityFile, err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else {
		var signers []ssh.Signer
		home, _ := os.UserHomeDir()
		defaultKeyPaths := []string{
			filepath.Join(home, ".ssh", "id_rsa"),
			// filepath.Join(home, ".ssh", "id_ed25519"),
			// filepath.Join(home, ".ssh", "id_dsa"),
		}
		for _, keyPath := range defaultKeyPaths {
			key, err := os.ReadFile(keyPath)
			if err == nil {
				signer, err := ssh.ParsePrivateKey(key)
				if err == nil {
					signers = append(signers, signer)
				}
			}
		}
		if len(signers) > 0 {
			logger.Debug("auth: found %d public key(s) in default locations", len(signers))
			authMethods = append(authMethods, ssh.PublicKeys(signers...))
		}

		challenge := func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			logger.Debug("auth: keyboard-interactive challenge. Questions: %v", questions)
			answers = make([]string, len(questions))
			for i, q := range questions {
				q = strings.TrimSpace(q)
				fmt.Printf("%s ", q)
				isPassword := strings.Contains(strings.ToLower(q), "password")
				isMfa := strings.Contains(strings.ToLower(q), "verification code") || strings.Contains(strings.ToLower(q), "otp")

				if isPassword {
					var password string = conn.GetPassword()
					if password == "" {
						logger.Debug("auth: prompting for password via keyboard-interactive\n")
						password, err = readPassword("Password: ")
						if err != nil {
							return nil, err
						}
						usedPassword = password
					} else {
						logger.Debug("auth: using stored password for keyboard-interactive\n")
					}
					answers[i] = password
				} else if isMfa {
					var mfaSecret string = conn.GetMfaSecret()
					if mfaSecret == "" {
						logger.Debug("auth: prompting for MFA secret via keyboard-interactive\n")
						mfaSecret, err = readPassword("MFA Secret: ")
						if err != nil {
							return nil, err
						}
						usedMfaSecret = mfaSecret
					} else {
						logger.Debug("auth: using stored MFA secret for keyboard-interactive\n")
					}
					otp, err := km.GenerateOTP(mfaSecret)
					if err != nil {
						return nil, fmt.Errorf("failed to generate OTP: %w", err)
					}
					answers[i] = otp
				} else {
					logger.Debug("auth: generic prompt in keyboard-interactive: %s", q)
					ans, err := bufio.NewReader(os.Stdin).ReadString('\n')
					if err != nil {
						return nil, err
					}
					answers[i] = strings.TrimSpace(ans)
				}
			}
			return answers, nil
		}
		authMethods = append(authMethods, ssh.KeyboardInteractive(challenge))

		authMethods = append(authMethods, ssh.PasswordCallback(func() (secret string, err error) {
			password := conn.GetPassword()
			if password == "" {
				logger.Debug("auth: prompting for password via password-callback")
				password, err = readPassword("Password: ")
				if err != nil {
					return "", err
				}
				usedPassword = password
			} else {
				logger.Debug("auth: using stored password for password-callback")
			}
			return password, nil
		}))
	}

	hostKeyCallback, err := createHostKeyCallback()
	if err != nil {
		return nil, conn, fmt.Errorf("failed to create host key callback: %w", err)
	}

	logger.Debug("auth: offering %d method(s)", len(authMethods))
	clientConfig := &ssh.ClientConfig{
		User:            conn.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", conn.HostName, conn.Port)
	logger.Debug("ssh: dialing %s with user %s", addr, conn.User)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, conn, fmt.Errorf("failed to dial: %w", err)
	}

	go savePwd(conn, usedMfaSecret, usedPassword)
	return client, conn, nil
}

// RunCommand executes a non-interactive command on a remote host.
// It streams stdout/stderr and returns the command's exit code.
func RunCommand(hostAlias string, command string, scriptContent io.Reader) (int, error) {
	client, _, err := EstablishSSHClient(hostAlias)
	if err != nil {
		return 1, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return 1, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if scriptContent != nil {
		session.Stdin = scriptContent
		if command == "" {
			// If we are running a script, and no command is specified,
			// we should explicitly run a shell to interpret the script.
			command = "/bin/sh"
		}
	}

	// 去除前后 单引号
	// command = strings.Trim(command, "'")
	if strings.HasPrefix(command, "'") && strings.HasSuffix(command, "'") {
		command = command[1 : len(command)-1]
	}

	err = session.Run(command)
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			// Command ran and exited with a non-zero status. This is not a
			// library error. We return the exit code.
			return exitErr.ExitStatus(), nil
		}
		// Other errors (network, etc.)
		return 1, fmt.Errorf("failed to run command: %w", err)
	}

	// Command ran successfully.
	return 0, nil
}

// RunSSH establishes an interactive SSH connection using the native Go SSH library.
// It handles public key, password, and keyboard-interactive (MFA/OTP) authentication.
func RunSSH(sshCmd string, conn config.SshConfigItem, args []string) error {
	client, _, err := establishSSHClient(conn.Host, &conn)
	if err != nil {
		return err
	}
	defer client.Close()

	// Create session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Set up terminal
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to make terminal raw: %w", err)
	}
	defer term.Restore(fd, oldState)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := term.GetSize(fd)
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %w", err)
	}

	// Request PTY
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, ssh.TerminalModes{}); err != nil {
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	// Handle window size changes
	watchWindowSize(fd, session)

	// Start shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// Wait for the session to finish
	err = session.Wait()
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			// The program has exited with an exit code != 0.
			// This is not a problem with the connection itself.
			// We can just exit with the same code.
			// os.Exit(exitErr.ExitStatus())
			return exitErr
		}
		// It's some other error, like the connection broke.
		return err
	}

	return nil
}

func createHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	knownHostsPath := fmt.Sprintf("%s/.ssh/known_hosts", home)

	// Ensure the known_hosts file exists
	f, err := os.OpenFile(knownHostsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open known_hosts file: %w", err)
	}
	f.Close()

	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create knownhosts callback: %w", err)
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := hostKeyCallback(hostname, remote, key)
		if err == nil {
			return nil
		}

		keyErr, ok := err.(*knownhosts.KeyError)
		if !ok {
			return fmt.Errorf("unexpected host key error: %w", err)
		}

		// Host is not in known_hosts, add it
		if len(keyErr.Want) == 0 {
			logger.Debug("Adding new host %s to known_hosts", hostname)
			f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				return fmt.Errorf("could not write to known_hosts file: %v", err)
			}
			defer f.Close()
			_, err = f.WriteString(knownhosts.Line([]string{hostname}, key) + "\n")
			return err
		}

		// Host key mismatch. Remove old key and add new one.
		logger.Debug("WARNING: Host key mismatch for %s. Removing old key and adding new one.", hostname)
		logger.Debug("Old key fingerprint(s): %s", keyErr.Want[0].Key.Type())
		logger.Debug("New key fingerprint: %s", ssh.FingerprintSHA256(key))

		// Use ssh-keygen to remove the old key
		cmd := exec.Command("ssh-keygen", "-R", hostname)
		if err := cmd.Run(); err != nil {
			logger.Debug("WARNING: failed to remove old host key with ssh-keygen: %v", err)
			// Fallback to manual removal attempt
			if err := removeHostFromKnownHostsManual(knownHostsPath, hostname); err != nil {
				return fmt.Errorf("failed to remove old host key for %s: %w", hostname, err)
			}
		}

		// Add the new key
		f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("could not write to known_hosts file: %v", err)
		}
		defer f.Close()
		_, err = f.WriteString(knownhosts.Line([]string{hostname}, key) + "\n")
		return err
	}, nil
}

// removeHostFromKnownHostsManual provides a basic way to remove a host from known_hosts
// This is a fallback if `ssh-keygen -R` fails.
func removeHostFromKnownHostsManual(knownHostsPath, hostname string) error {
	in, err := os.ReadFile(knownHostsPath)
	if err != nil {
		return err
	}

	var out []byte
	for len(in) > 0 {
		_, hosts, _, _, rest, err := ssh.ParseKnownHosts(in)
		if err != nil {
			// Handle parse errors, maybe just append the rest of the file
			out = append(out, in...)
			break
		}

		// Check if the current line matches the hostname to be removed
		isMatch := false
		for _, h := range hosts {
			if h == hostname {
				isMatch = true
				break
			}
		}

		if !isMatch {
			// Keep the line
			line := in[:len(in)-len(rest)]
			out = append(out, line...)
		}
		in = rest
	}

	return os.WriteFile(knownHostsPath, out, 0600)
}

func savePwd(connConf config.SshConfigItem, otpSecret, inputPassword string) {
	var updated bool
	if inputPassword != "" {
		connConf.Password = inputPassword
		updated = true
	}
	if otpSecret != "" {
		connConf.MFASecret = otpSecret
		updated = true
	}

	if updated {
		config.SaveConfigFromConn(connConf)
	}
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}
