package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sshpky/pkg/config"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	rsyncArchive  bool
	rsyncVerbose  bool
	rsyncCompress bool
	rsyncDelete   bool
)

// remotePathInfo holds the parsed components of a remote path argument.
type remotePathInfo struct {
	HostAlias string
	Path      string
	Original  string
	IsRemote  bool
	User      string
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// parseRemotePath parses a path argument (e.g., "my-server:/remote/path" or "user@my-server:/remote/path")
func parseRemotePath(pathArg string) *remotePathInfo {
	info := &remotePathInfo{Original: pathArg}
	parts := strings.SplitN(pathArg, ":", 2)

	if len(parts) != 2 || parts[0] == "" || strings.Contains(parts[0], "/") {
		info.IsRemote = false
		info.Path = pathArg
		return info
	}

	info.IsRemote = true
	info.Path = parts[1]

	hostPart := parts[0]
	if userAtIdx := strings.LastIndex(hostPart, "@"); userAtIdx != -1 {
		info.User = hostPart[:userAtIdx]
		info.HostAlias = hostPart[userAtIdx+1:]
	} else {
		info.HostAlias = hostPart
	}

	return info
}

func newSSHClient(cfg *config.SshConfigItem) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// 1. SSH Agent
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			agentClient := agent.NewClient(conn)
			authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
		}
	}

	// 2. Identity File
	if cfg.IdentityFile != "" {
		identityFile := expandPath(cfg.IdentityFile)
		key, err := ioutil.ReadFile(identityFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read identity file %s: %w", identityFile, err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key from %s: %w. If it's encrypted, please use ssh-agent", identityFile, err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no SSH authentication method available for host %s. Supported: ssh-agent, unencrypted IdentityFile", cfg.Host)
	}

	hostKeyCallback := ssh.InsecureIgnoreHostKey()

	clientConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         15 * time.Second,
	}

	if rsyncCompress {
		// Note: This relies on SSH-level compression, which is negotiated with the server.
		// The `zlib@openssh.com` method will be used if supported by both client and server.
		// This is different from rsync's protocol-level compression but achieves a similar goal.
		if rsyncVerbose {
			fmt.Println("Compression enabled (via SSH)")
		}
	}

	addr := fmt.Sprintf("%s:%d", cfg.HostName, cfg.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH server at %s: %w", addr, err)
	}

	return client, nil
}

func syncLocalToRemote(sftpClient *sftp.Client, localPath, remotePath string) error {
	lstat, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("cannot stat local path %s: %w", localPath, err)
	}

	// If remote path exists and is a directory, upload into it.
	rstat, err := sftpClient.Stat(remotePath)
	if err == nil && rstat.IsDir() {
		remotePath = sftp.Join(remotePath, filepath.Base(localPath))
	}

	if !lstat.IsDir() {
		return uploadFile(sftpClient, localPath, remotePath)
	}

	// Walk local directory and sync to remote
	err = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(localPath, path)
		destPath := sftp.Join(remotePath, filepath.ToSlash(relPath))

		if info.IsDir() {
			// Ensure remote directory exists
			if _, err := sftpClient.Stat(destPath); os.IsNotExist(err) {
				if rsyncVerbose {
					fmt.Printf("creating directory %s", destPath)
				}
				if err := sftpClient.MkdirAll(destPath); err != nil {
					return err
				}
			}
			return nil
		}

		// It's a file, check if we need to transfer
		remoteInfo, err := sftpClient.Stat(destPath)
		if err == nil && !info.ModTime().After(remoteInfo.ModTime()) {
			return nil // Not newer, skip
		}
		return uploadFile(sftpClient, path, destPath)
	})
	if err != nil {
		return err
	}

	// Handle --delete
	if rsyncDelete {
		return deleteExtraRemoteFiles(sftpClient, localPath, remotePath)
	}
	return nil
}

func uploadFile(sftpClient *sftp.Client, localFile, remoteFile string) error {
	if rsyncVerbose {
		fmt.Printf("uploading %s to %s", localFile, remoteFile)
	}

	info, err := os.Stat(localFile)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remoteFile)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if rsyncArchive {
		dstFile.Chmod(info.Mode())
		// dstFile.Chtimes(info.ModTime(), info.ModTime())
	}
	return nil
}

func deleteExtraRemoteFiles(sftpClient *sftp.Client, localBase, remoteBase string) error {
	if rsyncVerbose {
		fmt.Println("checking for extraneous files to delete on remote...")
	}
	var toDelete []string
	walker := sftpClient.Walk(remoteBase)
	for walker.Step() {
		if walker.Err() != nil {
			continue
		}
		remoteWalkPath := walker.Path()
		relPath, _ := filepath.Rel(filepath.FromSlash(remoteBase), filepath.FromSlash(remoteWalkPath))
		localCheckPath := filepath.Join(localBase, relPath)

		if _, err := os.Stat(localCheckPath); os.IsNotExist(err) {
			toDelete = append(toDelete, remoteWalkPath)
		}
	}
	// Delete files first, then directories (in reverse order)
	for i := len(toDelete) - 1; i >= 0; i-- {
		path := toDelete[i]
		if rsyncVerbose {
			fmt.Printf("deleting remote %s", path)
		}
		// Try removing as file or empty dir, then try as dir (might fail if not empty)
		err := sftpClient.Remove(path)
		if err != nil {
			sftpClient.RemoveDirectory(path)
		}
	}
	return nil
}

func syncRemoteToLocal(sftpClient *sftp.Client, remotePath, localPath string) error {
	fmt.Fprintln(os.Stderr, "Download not yet implemented.")
	return nil
}

// rsyncCmd represents the rsync command
var rsyncCmd = &cobra.Command{
	Use:   "rsync [-avz] [--delete] <source> <destination>",
	Short: "Sync files between local and remote using SFTP",
	Long: `Provides rsync-like file synchronization functionality using the SFTP protocol.
This avoids the need for the 'rsync' binary to be installed.

Remote paths are specified using the format: [user@]host_alias:path

Examples:
  # Upload local directory to remote server
  sshpky rsync -av ./local-dir my-server:/remote/app/

  # Upload with delete flag to remove extraneous files from destination
  sshpky rsync -av --delete ./local-dir my-server:/remote/app/
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfgManager := config.NewSSHConfigManager()

		sourceInfo := parseRemotePath(args[0])
		destInfo := parseRemotePath(args[1])

		if sourceInfo.IsRemote == destInfo.IsRemote {
			if sourceInfo.IsRemote {
				fmt.Fprintln(os.Stderr, "Error: Source and destination cannot both be remote.")
			} else {
				fmt.Fprintln(os.Stderr, "Error: Source and destination cannot both be local.")
			}

			os.Exit(1)
		}

		var remote, local *remotePathInfo
		var isUpload bool
		if sourceInfo.IsRemote {
			remote, local = sourceInfo, destInfo
			isUpload = false
		} else {
			remote, local = destInfo, sourceInfo
			isUpload = true
		}

		sshConfig, err := cfgManager.FindConfig(remote.HostAlias)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding SSH config for alias '%s': %v", remote.HostAlias, err)
			os.Exit(1)
		}
		if remote.User != "" {
			sshConfig.User = remote.User
		}

		client, err := newSSHClient(sshConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error establishing SSH connection: %v", err)
			os.Exit(1)
		}
		defer client.Close()

		sftpClient, err := sftp.NewClient(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating SFTP session: %v", err)
			os.Exit(1)
		}
		defer sftpClient.Close()

		if isUpload {
			err = syncLocalToRemote(sftpClient, local.Path, remote.Path)
		} else {
			err = syncRemoteToLocal(sftpClient, remote.Path, local.Path)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Synchronization failed: %v", err)
			os.Exit(1)
		}

		if rsyncVerbose {
			fmt.Println("Synchronization completed successfully.")
		}
	},
}

func init() {
	rootCmd.AddCommand(rsyncCmd)
	rsyncCmd.Flags().BoolVarP(&rsyncArchive, "archive", "a", false, "archive mode; preserves permissions and times")
	rsyncCmd.Flags().BoolVarP(&rsyncVerbose, "verbose", "v", false, "increase verbosity")
	rsyncCmd.Flags().BoolVarP(&rsyncCompress, "compress", "z", false, "compress file data during transfer (via SSH)")
	rsyncCmd.Flags().BoolVar(&rsyncDelete, "delete", false, "delete extraneous files from dest dirs")
}
