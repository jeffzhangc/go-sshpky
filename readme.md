# Go-sshpky

[![Go-sshpky](https://github.com/zhanghailiang/go-sshpky/actions/workflows/go.yml/badge.svg)](https://github.com/zhanghailiang/go-sshpky/actions/workflows/go.yml)

`go-sshpky` is a command-line tool for securely managing and using SSH connections. It enhances security by integrating with system keychains, while providing a user-friendly interface for managing SSH hosts in logical groups.

## Features

*   **Flexible Credential Storage**: Supports multiple methods for storing sensitive data. Credentials can be stored securely in your OS's native keychain (`StoreKeyChain`) for maximum security, or encrypted in a local file (`StoreFile`) for portability or in environments without a system keychain.
*   **Group Management**: Organizes SSH hosts into groups (e.g., `production`, `staging`) for easy management.
*   **MFA/OTP Support**: Automatically handles Time-based One-Time Passwords (TOTP) for multi-factor authentication.
*   **Simplified SSH Workflow**: A clear and simple CLI for managing and connecting to your SSH hosts.
*   **Cross-Platform**: Supports macOS, Linux, and Windows.

## Installation

You can install `go-sshpky` using Go:

```bash
go install github.com/zhanghailiang/go-sshpky@latest
```

Ensure that `$GOPATH/bin` is in your system's `PATH`.

## Quick Start

The typical workflow follows a `group -> host -> connect` pattern.

### 1. Create and Select a Group

First, create a group to organize your hosts.

```bash
# Create a new group named 'work'
sshpky mg add work

# Set 'work' as the default group for subsequent commands
sshpky mg use work
```

### 2. Add a New SSH Host

Now, add a new host configuration to the default group (`work`). The `ms add` command will interactively prompt you for connection details.

```bash
# Add a new host configuration
sshpky ms add
```

You will be prompted to enter:
- A name for the host (e.g., `web-server-1`).
- The SSH user and hostname (`user@host`).
- The authentication method and credentials (password, private key, etc.).

### 3. Connect to a Host

Once a host is added, you can connect to it using its name.

```bash
# Connect to the host you just added
sshpky conn web-server-1
```

You can also connect by specifying the user and host directly, which will use the stored configuration if a match is found.

```bash
sshpky conn user@host
```

## Documentation

For more detailed instructions, including how to manage configurations, use MFA, and more, please refer to our [full Usage Guide](./docs/usage/guide.md).

## Testing

This project includes a Docker-based environment for testing various SSH authentication methods. For detailed instructions on how to run the test suite, please see [TESTING.md](./TESTING.md).

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
