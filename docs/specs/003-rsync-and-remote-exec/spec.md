# Spec 003: Rsync and Remote Exec Commands

## 1. Background

The `go-sshpky` tool currently provides robust SSH key management and connection capabilities. However, users lack integrated, simplified methods for two common remote operations: file synchronization and remote command execution. They currently have to switch to separate tools like `rsync` and `ssh`, re-configuring authentication and host details, which is inefficient and fragments the user experience. This proposal aims to integrate these functionalities directly into `go-sshpky`.

## 2. Goals / Non-goals

### Goals

*   Introduce a new `rsync` subcommand for simplified file/directory synchronization.
*   Introduce a new `exec` subcommand for remote command and script execution.
*   Reuse `go-sshpky`'s existing host alias and authentication configuration for a seamless user experience.
*   Propagate remote exit codes to the local process to support scripting and automation.
*   Provide real-time feedback (stdout/stderr) for remote commands.

### Non-goals

*   Replicating the entire feature set of the standard `rsync` tool. We will only implement the most critical flags initially.
*   Introducing new authentication mechanisms. All auth will leverage the existing framework.
*   Building a complex remote session management system. The `exec` command is for stateless, one-off executions.

## 3. Use Cases

*   **UC-1 (File Sync):** A developer wants to quickly upload their latest local build artifacts to a remote development server defined in their `go-sshpky` config.
*   **UC-2 (Remote Script):** A sysadmin needs to run a maintenance script (stored locally) on multiple servers managed by `go-sshpky`.
*   **UC-3 (Automation):** An automated CI/CD pipeline script uses the `exec` command to run a deployment command on a production server and needs to know if the command succeeded or failed based on its exit code.
*   **UC-4 (Quick Check):** A user wants to quickly check the disk space (`df -h`) on a remote server without initiating a full interactive SSH session.

## 4. Requirements

### Functional Requirements (FR)

*   **FR-1 (`rsync` subcommand):**
    *   MUST be initiated via `go-sshpky rsync ...`.
    *   MUST support both upload (local to remote) and download (remote to local) synchronization.
    *   MUST use `<host_alias>:<path>` syntax to specify remote locations.
    *   MUST support the following flags:
        *   `-a, --archive`: Archive mode (preserves permissions, ownership, etc.).
        *   `-v, --verbose`: Verbose output.
        *   `-z, --compress`: Enable compression.
        *   `--delete`: Delete extraneous files from the destination directory.
*   **FR-2 (`exec` subcommand):**
    *   MUST be initiated via `go-sshpky exec ...`.
    *   MUST support executing an inline command string on a specified host alias.
    *   MUST use `--` to separate the command string from the tool's own arguments.
    *   MUST support executing a local script file on the remote host via a `-f, --file` flag.
*   **FR-3 (Authentication):**
    *   Both commands MUST use the host alias configuration from `go-sshpky` to handle authentication (keys, passwords, etc.).

### Non-Functional Requirements (NFR)

*   **NFR-1 (Usability):** The command syntax MUST be clear and simple.
*   **NFR-2 (Performance):** Remote command output MUST be streamed in real-time to the local terminal.
*   **NFR-3 (Integration):** Remote command exit codes MUST be propagated as the exit code of the `go-sshpky` process.

## 5. Proposed Solution

The recommended solution is to implement the two new subcommands (`rsync` and `exec`) using the Cobra library, consistent with the existing command structure.

*   For the `rsync` command, a native Go implementation will be developed to provide rsync-like functionality without depending on an external `rsync` binary. This will be achieved using the SFTP protocol over the established SSH connection. A Go SFTP library (e.g., `github.com/pkg/sftp`) will be used to programmatically list directories, compare file metadata (modification time, size), and transfer files. This approach ensures `go-sshpky` is self-contained and allows for a custom implementation of core features like incremental updates and deletions.

*   For the `exec` command, we will use the `golang.org/x/crypto/ssh` library, which is likely already a dependency. The command will establish an SSH session, execute the provided command or script, and pipe the remote `stdout` and `stderr` directly to the local process's `os.Stdout` and `os.Stderr` to achieve real-time streaming. The exit code from the remote command will be captured and used for the local process's exit code.

## 6. Architecture & Data Flow

1.  User runs `go-sshpky rsync|exec <host_alias> ...`.
2.  Cobra framework parses the command, flags, and arguments.
3.  The command's handler retrieves the host configuration (user, address, auth method) for `<host_alias>` using existing config management logic in `pkg/config`.
4.  An SSH connection is established using the retrieved credentials.
    *   **For `exec`:** An SSH channel is opened, the command is executed, I/O is streamed, and the exit code is captured.
    *   **For `rsync`:** An SFTP session is initiated over the SSH connection. The native Go implementation will then perform directory traversal, file metadata comparison, and file transfers using the SFTP session.
5.  Errors (connection, execution) are reported to the user.
6.  The `go-sshpky` process exits with the appropriate exit code (0 for success, remote's exit code on failure, or a custom error code for connection issues).

## 7. Interfaces / Data Model

### CLI Interface

```
# Rsync command
go-sshpky rsync [flags] <source> <destination>

# Flags: -a, -v, -z, --delete

# Example (upload):
go-sshpky rsync -avz ./local-dir my-server:/remote/app/

# Example (download):
go-sshpky rsync -avz my-server:/remote/app/ ./local-dir

# Exec command (inline)
go-sshpky exec <host_alias> [flags] -- '<command>'

# Example (inline):
go-sshpky exec my-server -- 'ls -la /var/www'

# Exec command (script)
go-sshpky exec <host_alias> -f <local_script_path>

# Example (script):
go-sshpky exec my-server -f ./deploy.sh
```

## 8. Error Handling & Edge Cases

*   **Invalid Host Alias:** If `<host_alias>` is not found in the config, print an error and exit with a non-zero status code.
*   **Authentication Failure:** Report auth failure clearly and exit.
*   **Connection Failure:** Report network or SSH handshake errors and exit.
*   **Remote Command Failure (`exec`):** The error will be visible in the streamed stderr. The process will exit with the remote command's exit code.
*   **`rsync` Failure:** Errors from the native SFTP implementation (e.g., permission denied, I/O errors) will be reported to the user. The process will exit with a non-zero status code.

## 9. Rollout / Migration / Rollback

*   **Rollout:** The new commands will be added in a new version of the `go-sshpky` binary. This is a purely additive change with no impact on existing functionality.
*   **Migration:** No data migration is needed.
*   **Rollback:** If issues arise, a release can be made that removes the new subcommands.

## 10. Work Breakdown

1.  **Task 1 (Setup):** Create new command files `cmd/rsync.go` and `cmd/exec.go` using Cobra structure.
2.  **Task 2 (`exec` implementation):** Implement the `exec` command logic using the `golang.org/x/crypto/ssh` package. Add support for inline commands and the `--file` flag. Ensure I/O streaming and exit code propagation.
3.  **Task 3 (`rsync` implementation):** Implement the native `rsync` command logic using a Go SFTP library. This includes directory traversal, file metadata comparison, handling uploads/downloads, and implementing flags like `--delete`.
4.  **Task 4 (Testing):** Add unit and integration tests for both commands, covering success paths, argument parsing, and error conditions.
5.  **Task 5 (Documentation):** Update `README.md` and any relevant CLI help text to document the new commands.

## 11. Testing Notes

*   Test `rsync` for both upload and download.
*   Test `rsync` with all specified flags (`-a`, `-v`, `-z`, `--delete`).
*   Test `exec` with both simple inline commands and multi-line local scripts.
*   Test exit code propagation for both successful (`0`) and failed (non-zero) remote commands.
*   Test authentication against different server types if possible (e.g., key-based, password-based).
*   Test error handling for invalid host aliases and network connection failures.

## 12. Acceptance Criteria

*   [ ] `go-sshpky rsync my-server:/path/to/remote /local/path` successfully downloads the file.
*   [ ] `go-sshpky rsync /local/path my-server:/path/to/remote` successfully uploads the file.
*   [ ] `go-sshpky exec my-server -- 'echo "hello"'` prints "hello" to the console and exits with code 0.
*   [ ] `go-sshpky exec my-server -- 'exit 123'` exits with code 123.
*   [ ] `go-sshpky exec my-server -f ./test.sh` successfully executes the script on the remote server.
*   [ ] Running a command against a non-existent alias fails with a clear error message.

## 13. Open Questions

*   None at this time.

## 14. Future Work

*   Add support for more `rsync` flags.
*   Add progress bar/indicator for `rsync` transfers.
*   Allow passing arguments to scripts executed with `exec -f`.
