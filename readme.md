# Go-sshpky

[![Go-sshpky](https://github.com/zhanghailiang/go-sshpky/actions/workflows/go.yml/badge.svg)](https://github.com/zhanghailiang/go-sshpky/actions/workflows/go.yml)

`go-sshpky` is a command-line tool for securely managing and using SSH private keys. It enhances security by integrating with system keychains and supporting hardware-based keys, while providing a user-friendly interface for managing SSH connections.

## Features

*   **Secure Key Management**: Integrates with system keychains (macOS, Windows, Linux) to store private keys securely.
*   **MFA/OTP Support**: Automatically handles Time-based One-Time Passwords (TOTP) for multi-factor authentication.
*   **Simplified SSH Config**: Manages SSH configurations for multiple hosts with grouping.
*   **User-Friendly CLI**: Provides an intuitive interface for key and connection management.
*   **Cross-Platform**: Supports macOS, Linux, and Windows.

## Installation

You can install `go-sshpky` using Go:

```bash
go install github.com/zhanghailiang/go-sshpky@latest
```

Ensure that `$GOPATH/bin` is in your system's `PATH`.

## Quick Start

### 1. Initialize Configuration

First, initialize the configuration file. This will create a default `~/.go-sshpky/config.yaml` file.

```bash
go-sshpky init
```

### 2. Add a New SSH Host

Use the `add` command to add a new host to your configuration. This will guide you through the process of setting up the connection details.

```bash
go-sshpky add
```

You will be prompted to enter:
- A name for the host.
- The SSH user and hostname (`user@host`).
- The authentication method (e.g., password, private key, hardware key).

### 3. Connect to a Host

Once a host is added, you can connect to it using the `conn` command.

```bash
go-sshpky conn <host-name>
```

### 4. List Hosts

To see a list of all configured hosts, use the `list` command.

```bash
go-sshpky list
```

## Documentation

For more detailed information, advanced usage, and architectural details, please refer to our [full documentation](./docs/README.md).

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
