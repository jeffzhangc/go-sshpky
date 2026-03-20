package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sshpky/pkg/config"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

var (
	testDir       string
	testSSHConfig string
)

func TestMain(m *testing.M) {
	// Setup
	var err error
	testDir, err = os.MkdirTemp("", "sshpky_tests")
	if err != nil {
		log.Fatalf("Failed to create temp dir for testing: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Get absolute path for test identity file
	identityFile, err := filepath.Abs("../ssh_test_servers/test_id_rsa")
	if err != nil {
		log.Fatalf("Failed to get absolute path for identity file: %v", err)
	}

	// --- Docker Compose Setup ---
	composeFile := "../ssh_test_servers/docker-compose.yml"
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		log.Fatalf("docker-compose file not found: %s", composeFile)
	}

	// Check if services are already running
	psCmd := exec.Command("docker-compose", "-f", composeFile, "ps", "-q")
	psOutput, err := psCmd.Output()
	if err != nil {
		log.Fatalf("Failed to check docker-compose status: %v", err)
	}

	if len(strings.TrimSpace(string(psOutput))) == 0 {
		fmt.Println("Starting docker-compose services for testing...")
		upCmd := exec.Command("docker-compose", "-f", composeFile, "up", "-d", "--build")
		output, err := upCmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to start docker-compose services: %v\nOutput: %s", err, string(output))
		}
	} else {
		fmt.Println("Docker-compose services already running.")
	}

	// Teardown function for docker-compose
	defer func() {
		fmt.Println("Stopping docker-compose services...")
		downCmd := exec.Command("docker-compose", "-f", composeFile, "down")
		output, err := downCmd.CombinedOutput()
		if err != nil {
			log.Printf("Failed to stop docker-compose services: %v\nOutput: %s", err, string(output))
		}
	}()

	// It can take a moment for the ssh server to be ready.
	log.Println("Waiting for SSH servers to become available...")
	time.Sleep(5 * time.Second) // Increased sleep time for more reliability
	log.Println("SSH servers should be up.")

	testSSHConfig = filepath.Join(testDir, "config")
	configContent := fmt.Sprintf(`
Host t_pubkey
    HostName 127.0.0.1
    Port 2223
    User testuser
    IdentityFile %s
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null

Host t_pass
    HostName 127.0.0.1
    Port 2222
    User testuser
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null

Host t_mfa
    HostName 127.0.0.1
    Port 2224
    User testuser
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
`, identityFile)

	err = os.WriteFile(testSSHConfig, []byte(configContent), 0600)
	if err != nil {
		log.Fatalf("Failed to write test ssh config: %v", err)
	}

	// Point the config manager to the test config
	config.SetConfigPath(testSSHConfig)
	defer config.ResetConfigManager()

	// Run the tests
	exitCode := m.Run()

	os.Exit(exitCode)
}

// executeCommand is a helper that executes a cobra command and captures its output
// and exit code. It's essential for testing commands that call os.Exit.
func executeCommand(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	// Reset flag values to their defaults before running the command.
	// This prevents state from leaking between tests.
	execCmd.Flags().Visit(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue)
	})

	// This function captures the exit code by replacing the os.Exit function.
	var exitCodeContainer int
	var once sync.Once
	done := make(chan struct{})

	originalOsExit := osExit
	osExit = func(code int) {
		exitCodeContainer = code
		once.Do(func() { close(done) })
	}
	defer func() { osExit = originalOsExit }()

	// Redirect stdout and stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	os.Stdout = wOut
	os.Stderr = wErr
	defer func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	// Capture output in a separate goroutine
	var outBuf, errBuf bytes.Buffer
	var wgRead sync.WaitGroup
	wgRead.Add(2)

	go func() {
		io.Copy(&outBuf, rOut)
		wgRead.Done()
	}()
	go func() {
		io.Copy(&errBuf, rErr)
		wgRead.Done()
	}()

	// Execute the command in a goroutine so we can wait for completion
	go func() {
		defer once.Do(func() { close(done) })
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			// If Execute returns an error, it means os.Exit was not called.
			// This can happen on input validation errors from cobra.
			// We set a non-zero exit code.
			if exitCodeContainer == 0 {
				exitCodeContainer = 1
			}
			errBuf.WriteString(err.Error())
		}
	}()

	<-done // Wait for os.Exit to be called or command to finish

	wOut.Close()
	wErr.Close()
	wgRead.Wait() // Wait for I/O copy to finish

	return outBuf.String(), errBuf.String(), exitCodeContainer
}
