package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestExecCommand_Simple(t *testing.T) {
	stdout, stderr, exitCode := executeCommand(t, "exec", "t_pubkey", "--", "echo", "hello from test")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "hello from test") {
		t.Errorf("Expected stdout to contain 'hello from test', but it didn't. Stdout: %s", stdout)
	}
}

func TestExecCommand_Failing(t *testing.T) {
	_, stderr, exitCode := executeCommand(t, "exec", "t_pubkey", "--", "/bin/false")

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d. Stderr: %s", exitCode, stderr)
	}
}

func TestExecCommand_ExitCodePropagation(t *testing.T) {
	_, stderr, exitCode := executeCommand(t, "exec", "t_pubkey", "--", "sh", "-c 'exit 42'")

	if exitCode != 42 {
		t.Errorf("Expected exit code 42, got %d. Stderr: %s", exitCode, stderr)
	}
}

func TestExecCommand_ScriptFile(t *testing.T) {
	scriptContent := `#!/bin/sh
echo "hello from script file"
exit 0
`
	scriptFile, err := os.CreateTemp(testDir, "test_script_*.sh")
	if err != nil {
		t.Fatalf("Failed to create temp script file: %v", err)
	}
	defer os.Remove(scriptFile.Name())

	if _, err := scriptFile.WriteString(scriptContent); err != nil {
		t.Fatalf("Failed to write to temp script file: %v", err)
	}
	scriptFile.Close()
	stdout, stderr, exitCode := executeCommand(t, "exec", "t_pubkey", "-f", scriptFile.Name())

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "hello from script file") {
		t.Errorf("Expected stdout to contain 'hello from script file', but it didn't. Stdout: %s", stdout)
	}
}

func TestExecCommand_InvalidHost(t *testing.T) {
	_, stderr, exitCode := executeCommand(t, "exec", "non_existent_host", "--", "echo", "hello")

	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code for invalid host, got 0")
	}
	if !strings.Contains(stderr, "failed to find config") {
		t.Errorf("Expected stderr to contain 'failed to find config', but it didn't. Stderr: %s", stderr)
	}
}
