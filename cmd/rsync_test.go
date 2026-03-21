package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sshpky/pkg/config"
	sftpPkg "sshpky/pkg/sftp"
	"sshpky/pkg/sshrunner"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func TestParseRemotePath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantIsRemote bool
		wantHost     string
		wantUser     string
		wantPath     string
	}{
		{"local path with dot", "./local", false, "", "", "./local"},
		{"local path relative", "../local", false, "", "", "../local"},
		{"local path absolute", "/local/path", false, "", "", "/local/path"},
		// {"windows-style local path", "C:\\Users\\Test", false, "", "", "C:\\Users\\Test"},
		{"remote path", "host:/path", true, "host", "", "/path"},
		{"remote path with user", "user@host:/path/to/dir", true, "host", "user", "/path/to/dir"},
		{"remote path root", "host:/", true, "host", "", "/"},
		{"malformed no colon", "hostpath", false, "", "", "hostpath"},
		{"malformed empty host", ":/path", false, "", "", ":/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := sftpPkg.ParseRemotePath(tt.path)
			if info.IsRemote != tt.wantIsRemote {
				t.Errorf("parseRemotePath().IsRemote = %v, want %v", info.IsRemote, tt.wantIsRemote)
			}
			if info.HostAlias != tt.wantHost {
				t.Errorf("parseRemotePath().HostAlias = %v, want %v", info.HostAlias, tt.wantHost)
			}
			if info.User != tt.wantUser {
				t.Errorf("parseRemotePath().User = %v, want %v", info.User, tt.wantUser)
			}
			if info.Path != tt.wantPath {
				t.Errorf("parseRemotePath().Path = %v, want %v", info.Path, tt.wantPath)
			}
		})
	}
}

func TestRsyncCommand_ArgValidation(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		_, stderr, exitCode := executeCommand(t, "rsync")
		if exitCode == 0 {
			t.Error("Expected non-zero exit code, got 0")
		}
		if !strings.Contains(stderr, "accepts 2 arg") {
			t.Errorf("Expected stderr to contain 'accepts 2 arg', got: %s", stderr)
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		_, stderr, exitCode := executeCommand(t, "rsync", "a", "b", "c")
		if exitCode == 0 {
			t.Error("Expected non-zero exit code, got 0")
		}
		if !strings.Contains(stderr, "accepts 2 arg") {
			t.Errorf("Expected stderr to contain 'accepts 2 arg', got: %s", stderr)
		}
	})

	t.Run("both local paths", func(t *testing.T) {
		_, stderr, exitCode := executeCommand(t, "rsync", "./local1", "./local2")
		if exitCode == 0 {
			t.Error("Expected non-zero exit code, got 0")
		}
		if !strings.Contains(stderr, "Source and destination cannot both be local") {
			t.Errorf("Expected specific error for both local paths, got: %s", stderr)
		}
	})

	t.Run("both remote paths", func(t *testing.T) {
		_, stderr, exitCode := executeCommand(t, "rsync", "t_pubkey:/a", "t_pubkey:/b")
		if exitCode == 0 {
			t.Error("Expected non-zero exit code, got 0")
		}
		if !strings.Contains(stderr, "Source and destination cannot both be remote") {
			t.Errorf("Expected specific error for both remote paths, got: %s", stderr)
		}
	})

	t.Run("invalid host", func(t *testing.T) {
		_, stderr, exitCode := executeCommand(t, "rsync", "invalid-host:/src", "./dest")
		if exitCode == 0 {
			t.Error("Expected non-zero exit code, got 0")
		}
		if !strings.Contains(stderr, "Error finding SSH config for alias") {
			t.Errorf("Expected specific error for invalid host, got: %s", stderr)
		}
	})
}

func TestRsyncIntegration_Download(t *testing.T) {
	// Setup: Create remote source directory and files
	remoteBaseDir := fmt.Sprintf("/tmp/rsync-download-src-%d", time.Now().UnixNano())
	remoteSubdir := remoteBaseDir + "/subdir"
	// executeCommand(t, "exec", "t_pubkey", "--", fmt.Sprintf("rm -rf %s", remoteBaseDir))
	defer executeCommand(t, "exec", "t_pubkey", "--", fmt.Sprintf("rm -rf %s", remoteBaseDir))

	// Create remote directory structure
	executeCommand(t, "exec", "t_pubkey", "--", fmt.Sprintf("mkdir -p %s", remoteSubdir))
	// Create files using printf for better compatibility
	executeCommand(t, "exec", "t_pubkey", "--", "printf", "hello", ">", fmt.Sprintf("%s/file1.txt", remoteBaseDir))
	executeCommand(t, "exec", "t_pubkey", "--", "printf", "world", ">", fmt.Sprintf("%s/file2.txt", remoteSubdir))

	// 等待远程命令执行完成
	time.Sleep(500 * time.Millisecond)

	// Setup: Create local destination directory
	localDir, err := os.MkdirTemp("", "rsync-download-dest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(localDir)

	// Execute: Run the rsync download command
	// Note: Download to localDir, contents will be placed in localDir/rsync-download-src-xxx
	_, stderr, exitCode := executeCommand(t, "rsync", "-av", fmt.Sprintf("t_pubkey:%s", remoteBaseDir), localDir)
	if exitCode != 0 {
		t.Fatalf("Rsync download command failed with exit code %d. Stderr: %s", exitCode, stderr)
	}

	// The downloaded content will be in localDir with the base name of remote path
	downloadedDir := filepath.Join(localDir, filepath.Base(remoteBaseDir))

	// Verify: Check downloaded file1.txt
	file1Path := filepath.Join(downloadedDir, "file1.txt")
	content1, err := os.ReadFile(file1Path)
	if err != nil {
		t.Errorf("Failed to read downloaded file1.txt: %v", err)
	}
	if string(content1) != "hello" {
		t.Errorf("Content of file1.txt is '%s', want 'hello'", string(content1))
	}

	// Verify: Check downloaded subdir/file2.txt
	file2Path := filepath.Join(downloadedDir, "subdir", "file2.txt")
	content2, err := os.ReadFile(file2Path)
	if err != nil {
		t.Errorf("Failed to read downloaded file2.txt: %v", err)
	}
	if string(content2) != "world" {
		t.Errorf("Content of file2.txt is '%s', want 'world'", string(content2))
	}
}

func TestRsyncIntegration_DownloadDelete(t *testing.T) {
	// Setup: Create remote source directory with files
	remoteBaseDir := fmt.Sprintf("/tmp/rsync-dl-delete-src-%d", time.Now().UnixNano())
	defer executeCommand(t, "exec", "t_pubkey", "--", fmt.Sprintf("rm -rf %s", remoteBaseDir))

	out1, err1, code := executeCommand(t, "exec", "t_pubkey", "--", "mkdir", "-p", remoteBaseDir)
	fmt.Println(out1, err1, code)

	executeCommand(t, "exec", "t_pubkey", "--", "touch", fmt.Sprintf("%s/file1.txt", remoteBaseDir))
	executeCommand(t, "exec", "t_pubkey", "--", "touch", fmt.Sprintf("%s/to_delete.txt", remoteBaseDir))

	// 等待远程命令执行完成
	time.Sleep(500 * time.Millisecond)

	// Setup: Create local destination directory
	localDir, err := os.MkdirTemp("", "rsync-dl-delete-dest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(localDir)

	// Execute: First sync - download to localDir
	// Contents will be in localDir/<basename>
	_, stderr, exitCode := executeCommand(t, "rsync", "-a", fmt.Sprintf("t_pubkey:%s", remoteBaseDir), localDir)
	if exitCode != 0 {
		t.Fatalf("First rsync download failed. Stderr: %s", stderr)
	}

	// 等待同步完成
	time.Sleep(300 * time.Millisecond)

	// The downloaded content will be in localDir with the base name of remote path
	downloadedDir := filepath.Join(localDir, filepath.Base(remoteBaseDir))

	// Verify: Both files should exist
	if _, err := os.Stat(filepath.Join(downloadedDir, "to_delete.txt")); err != nil {
		t.Fatalf("to_delete.txt should exist after first sync: %v", err)
	}
	if _, err := os.Stat(filepath.Join(downloadedDir, "file1.txt")); err != nil {
		t.Fatalf("file1.txt should exist after first sync: %v", err)
	}

	// Setup: Delete a remote file
	executeCommand(t, "exec", "t_pubkey", "--", fmt.Sprintf("rm -rf %s/to_delete.txt", remoteBaseDir))

	// 等待远程删除完成
	time.Sleep(300 * time.Millisecond)

	// Execute: Second sync with --delete flag
	_, stderr, exitCode = executeCommand(t, "rsync", "-a", "--delete", fmt.Sprintf("t_pubkey:%s", remoteBaseDir), localDir)
	if exitCode != 0 {
		t.Fatalf("Rsync command with --delete failed. Stderr: %s", stderr)
	}

	// Verify: Check local state
	if _, err := os.Stat(filepath.Join(downloadedDir, "file1.txt")); err != nil {
		t.Errorf("file1.txt should still exist locally, but stat failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(downloadedDir, "to_delete.txt")); !os.IsNotExist(err) {
		t.Errorf("to_delete.txt should have been deleted locally, but it still exists")
	}
}

func TestRsyncIntegration_Upload(t *testing.T) {
	// Setup: Create local source directory and files
	localDir, err := os.MkdirTemp("", "rsync-local-src")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(localDir)

	if err := os.WriteFile(filepath.Join(localDir, "file1.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(localDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "subdir", "file2.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Setup: Define remote destination and ensure it's clean
	remoteBaseDir := fmt.Sprintf("/tmp/rsync-test-%d", time.Now().UnixNano())
	executeCommand(t, "exec", "t_pubkey", "--", fmt.Sprintf("rm -rf %s", remoteBaseDir))
	defer executeCommand(t, "exec", "t_pubkey", "--", fmt.Sprintf("rm -rf %s", remoteBaseDir))

	// Execute: Run the rsync upload command
	_, stderr, exitCode := executeCommand(t, "rsync", "-av", localDir, fmt.Sprintf("t_pubkey:%s", remoteBaseDir))
	if exitCode != 0 {
		t.Fatalf("Rsync command failed with exit code %d. Stderr: %s", exitCode, stderr)
	}

	// Verify: Use an SFTP client to check the remote state
	client, sftpClient := newSftpClient(t, "t_pubkey")
	defer client.Close()
	defer sftpClient.Close()

	// Verify file1.txt
	verifyRemoteFileContent(t, sftpClient, sftp.Join(remoteBaseDir, "file1.txt"), "hello")

	// Verify file2.txt in subdir
	verifyRemoteFileContent(t, sftpClient, sftp.Join(remoteBaseDir, "subdir", "file2.txt"), "world")
}

func TestRsyncIntegration_UploadDelete(t *testing.T) {
	// Setup: Create local source directory with files
	localDir, err := os.MkdirTemp("", "rsync-delete-src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(localDir)

	if err := os.WriteFile(filepath.Join(localDir, "file1.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	fileToDelete := filepath.Join(localDir, "to_delete.txt")
	if err := os.WriteFile(fileToDelete, []byte("delete me"), 0644); err != nil {
		t.Fatal(err)
	}

	// Setup: Define remote destination and ensure it's clean
	remoteDir := fmt.Sprintf("/tmp/rsync-delete-dest-%d", time.Now().UnixNano())
	remotePathForSync := remoteDir + "/" + filepath.Base(localDir)
	executeCommand(t, "exec", "t_pubkey", "--", fmt.Sprintf("rm -rf %s", remoteDir))
	defer executeCommand(t, "exec", "t_pubkey", "--", fmt.Sprintf("rm -rf %s", remoteDir))

	// Execute: First sync
	executeCommand(t, "rsync", "-a", localDir, fmt.Sprintf("t_pubkey:%s", remoteDir))

	// Setup: Delete a local file
	if err := os.Remove(fileToDelete); err != nil {
		t.Fatal(err)
	}

	// Execute: Second sync with --delete flag
	_, stderr, exitCode := executeCommand(t, "rsync", "-a", "--delete", localDir, fmt.Sprintf("t_pubkey:%s", remoteDir))
	if exitCode != 0 {
		t.Fatalf("Rsync command with --delete failed. Stderr: %s", stderr)
	}

	// Verify: Check remote state
	client, sftpClient := newSftpClient(t, "t_pubkey")
	defer client.Close()
	defer sftpClient.Close()

	if _, err := sftpClient.Stat(sftp.Join(remotePathForSync, "file1.txt")); err != nil {
		t.Errorf("file1.txt should still exist on remote, but stat failed: %v", err)
	}

	if _, err := sftpClient.Stat(sftp.Join(remotePathForSync, "to_delete.txt")); !os.IsNotExist(err) {
		t.Errorf("to_delete.txt should have been deleted from remote, but it still exists (err: %v)", err)
	}
}

// verifyRemoteFileContent is a test helper to read a remote file and check its content.
func verifyRemoteFileContent(t *testing.T, sftpClient *sftp.Client, path, wantContent string) {
	t.Helper()
	file, err := sftpClient.Open(path)
	if err != nil {
		t.Fatalf("Failed to open remote file %s: %v", path, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read content of remote file %s: %v", path, err)
	}

	if string(content) != wantContent {
		t.Errorf("Content of remote file %s is '%s', want '%s'", path, string(content), wantContent)
	}
}

// newSftpClient is a test helper to get an SFTP client for verification.
func newSftpClient(t *testing.T, hostAlias string) (*ssh.Client, *sftp.Client) {
	t.Helper()
	cfgManager := config.NewSSHConfigManager()
	sshConfig, err := cfgManager.FindConfig(hostAlias)
	if err != nil {
		t.Fatalf("Failed to find config for host '%s': %v", hostAlias, err)
	}

	client, _, err := sshrunner.EstablishSSHClient(sshConfig.Host)
	if err != nil {
		t.Fatalf("Failed to create ssh client for verification: %v", err)
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		t.Fatalf("Failed to create sftp client for verification: %v", err)
	}

	return client, sftpClient
}
