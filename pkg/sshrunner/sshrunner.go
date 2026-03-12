package sshrunner

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sshpky/pkg/config"
	"sshpky/pkg/km"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

// RunSSH establishes an interactive SSH connection using the native Go SSH library.
// It handles public key, password, and keyboard-interactive (MFA/OTP) authentication.
func RunSSH(sshCmd string, conn config.SshConfigItem, args []string) error {
	ms := config.NewSSHConfigManager()
	cnf, _ := ms.FindConfig(conn.Host)
	if cnf != nil {
		conn = *cnf
	}

	if conn.User == "" {
		currentUser, err := user.Current()
		if err == nil {
			log.Printf("ssh: user not specified, defaulting to current user %s", currentUser.Username)
			conn.User = currentUser.Username
		} else {
			log.Printf("ssh: warning: could not get current user: %v", err)
		}
	}

	var usedPassword, usedMfaSecret string
	authMethods := []ssh.AuthMethod{}

	// --- 1. Public Keys (Temporarily Disabled for Debugging) ---
	var signers []ssh.Signer
	// a. From config
	if conn.IdentityFile != "" {
		key, err := os.ReadFile(conn.IdentityFile)
		if err == nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				signers = append(signers, signer)
			}
		}
	} else {
		// b. From default locations
		home, _ := os.UserHomeDir()
		defaultKeyPaths := []string{
			filepath.Join(home, ".ssh", "id_rsa"),
			filepath.Join(home, ".ssh", "id_ed25519"),
			filepath.Join(home, ".ssh", "id_dsa"),
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
	}

	if len(signers) > 0 {
		log.Printf("auth: found %d public key(s)", len(signers))
		authMethods = append(authMethods, ssh.PublicKeys(signers...))
	}

	// --- 2. Keyboard Interactive ---
	// Handles MFA/OTP and can also handle password prompts from the server.
	challenge := func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
		log.Printf("auth: keyboard-interactive challenge received. Questions: %v", questions)
		answers = make([]string, len(questions))
		for i, q := range questions {
			q = strings.TrimSpace(q)
			fmt.Printf("%s ", q)

			isPassword := strings.Contains(strings.ToLower(q), "password")
			isMfa := strings.Contains(strings.ToLower(q), "verification code") || strings.Contains(strings.ToLower(q), "otp")

			if isPassword {
				var password string
				password = conn.GetPassword()
				if password == "" {
					log.Printf("auth: prompting user for password via keyboard-interactive\n")

					password, err = readPassword("Password: ")
					if err != nil {
						return nil, err
					}
					usedPassword = password // Capture for saving
				} else {
					log.Printf("auth: using stored password for keyboard-interactive\n")
				}
				answers[i] = password
			} else if isMfa {
				var mfaSecret string
				mfaSecret = conn.GetMfaSecret()
				if mfaSecret == "" {
					log.Printf("auth: prompting user for MFA secret via keyboard-interactive\n")
					mfaSecret, err = readPassword("MFA Secret: ")
					if err != nil {
						return nil, err
					}
					usedMfaSecret = mfaSecret // Capture for saving
				} else {
					log.Printf("auth: using stored MFA secret for keyboard-interactive\n")
				}
				otp, err := km.GenerateOTP(mfaSecret)
				if err != nil {
					return nil, fmt.Errorf("failed to generate OTP: %w", err)
				}
				answers[i] = otp
			} else {
				log.Printf("auth: generic prompt in keyboard-interactive: %s", q)
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
			log.Printf("auth: prompting user for password via keyboard-interactive")
			password, err = readPassword("Password: ")
			if err != nil {
				return "", err
			}
			usedPassword = password // Capture for saving
		} else {
			log.Printf("auth: using stored password for keyboard-interactive")
		}
		return password, nil
	}))

	// Setup host key callback
	hostKeyCallback, err := createHostKeyCallback()
	if err != nil {
		return fmt.Errorf("failed to create host key callback: %w", err)
	}

	log.Printf("auth: offering %d method(s)", len(authMethods))

	clientConfig := &ssh.ClientConfig{
		User:            conn.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         15 * time.Second,
	}

	// Dial
	addr := fmt.Sprintf("%s:%d", conn.HostName, conn.Port)
	log.Printf("ssh: dialing %s with user %s", addr, conn.User)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	// Save credentials if any new ones were used
	go savePwd(conn, usedMfaSecret, usedPassword)

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
			os.Exit(exitErr.ExitStatus())
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
			log.Printf("Adding new host %s to known_hosts", hostname)
			f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				return fmt.Errorf("could not write to known_hosts file: %v", err)
			}
			defer f.Close()
			_, err = f.WriteString(knownhosts.Line([]string{hostname}, key) + "\n")
			return err
		}

		// Host key mismatch. Remove old key and add new one.
		log.Printf("WARNING: Host key mismatch for %s. Removing old key and adding new one.", hostname)
		log.Printf("Old key fingerprint(s): %s", keyErr.Want[0].Key.Type())
		log.Printf("New key fingerprint: %s", ssh.FingerprintSHA256(key))

		// Use ssh-keygen to remove the old key
		cmd := exec.Command("ssh-keygen", "-R", hostname)
		if err := cmd.Run(); err != nil {
			log.Printf("WARNING: failed to remove old host key with ssh-keygen: %v", err)
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
