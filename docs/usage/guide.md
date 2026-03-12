# Go-sshpky Usage Guide

This guide provides comprehensive instructions on how to use `go-sshpky` for managing your SSH connections and keys.

## Table of Contents

1.  [Introduction](#1-introduction)
2.  [Installation](#2-installation)
3.  [Basic Usage](#3-basic-usage)
    *   [Initializing Configuration](#initializing-configuration)
    *   [Adding a New SSH Host](#adding-a-new-ssh-host)
    *   [Connecting to a Host](#connecting-to-a-host)
    *   [Listing Configured Hosts](#listing-configured-hosts)
4.  [Managing Groups](#4-managing-groups)
    *   [Listing Groups](#listing-groups)
    *   [Switching Active Group](#switching-active-group)
    *   [Adding a Host to a Specific Group](#adding-a-host-to-a-specific-group)
5.  [Managing SSH Config Items (`ms` command)](#5-managing-ssh-config-items-ms-command)
    *   [Viewing Config Items](#viewing-config-items)
    *   [Editing Config Items](#editing-config-items)
    *   [Deleting Config Items](#deleting-config-items)
6.  [Advanced Features](#6-advanced-features)
    *   [Multi-Factor Authentication (MFA/OTP)](#multi-factor-authentication-mfaotp)
    *   [Using Identity Files (Private Keys)](#using-identity-files-private-keys)
    *   [ProxyCommand for Jump Hosts](#proxycommand-for-jump-hosts)
7.  [Troubleshooting](#7-troubleshooting)
8.  [Examples](#8-examples)

---

## 1. Introduction

`go-sshpky` is a powerful command-line tool designed to simplify and secure your SSH workflow. It allows you to manage multiple SSH connections, credentials, and configurations efficiently, integrating with system keychains for enhanced security and supporting advanced features like MFA.

## 2. Installation

For installation instructions, please refer to the main [README.md](../../README.md) file.

## 3. Basic Usage

### Initializing Configuration

Before using `go-sshpky`, you need to initialize its configuration. This command creates a default `config.yaml` file in `~/.go-sshpky/`.

```bash
go-sshpky init
```

### Adding a New SSH Host

The `add` command interactively guides you through setting up a new SSH host entry.

```bash
go-sshpky add
```

You will be prompted for:
*   **Host Name**: A unique alias for your connection (e.g., `my-web-server`).
*   **User@Host**: The actual SSH user and hostname/IP (e.g., `ubuntu@192.168.1.100`).
*   **Port**: The SSH port (defaults to 22).
*   **Authentication Method**: Choose between password, private key, or hardware key.
*   **Password/Key Path/MFA Secret**: Depending on your chosen method.

### Connecting to a Host

Once a host is configured, you can connect to it using the `conn` command followed by the host name.

```bash
go-sshpky conn <host-name>
```

Example:
```bash
go-sshpky conn my-web-server
```

You can also pass SSH options directly:
```bash
go-sshpky conn -u admin -p 2222 example.com
```

### Listing Configured Hosts

To view all your configured SSH hosts, use the `list` command.

```bash
go-sshpky list
```

This will display a table of your hosts, their users, hostnames, and associated groups.

## 4. Managing Groups

`go-sshpky` allows you to organize your SSH configurations into groups (e.g., `production`, `staging`, `development`).

### Listing Groups

To see all defined groups:

```bash
go-sshpky mg list
```

### Switching Active Group

You can set an active group, which will be used for subsequent `add` or `conn` commands unless overridden.

```bash
go-sshpky mg use <group-name>
```

Example:
```bash
go-sshpky mg use production
```

### Adding a Host to a Specific Group

When adding a host, you can specify the group using the `-g` flag:

```bash
go-sshpky add -g development
```

Or connect to a host within a specific group:

```bash
go-sshpky conn -g staging appserver
```

## 5. Managing SSH Config Items (`ms` command)

The `ms` command provides more granular control over individual SSH configuration items.

### Viewing Config Items

To view config items for the current active group:

```bash
go-sshpky ms
```

To view config items for a specific group:

```bash
go-sshpky ms <group-name>
```

### Editing Config Items

The `ms` command will typically open an interactive editor (like `vi` or `nano`) to allow you to modify the configuration details for a host.

```bash
go-sshpky ms <host-name>
```

### Deleting Config Items

(Specific command for deletion to be added/confirmed based on `go-sshpky` CLI)

## 6. Advanced Features

### Multi-Factor Authentication (MFA/OTP)

`go-sshpky` can store and use TOTP secrets for MFA. When adding a host, if you provide an MFA secret, `go-sshpky` will automatically generate and attempt to use the OTP during connection.

### Using Identity Files (Private Keys)

You can specify an identity file (private key) for authentication when adding a host. `go-sshpky` will manage the path to this key.

```bash
go-sshpky add --identity /path/to/your/private_key
```

### ProxyCommand for Jump Hosts

`go-sshpky` supports `ProxyCommand` for connecting through jump hosts. You can specify this when adding or editing a host configuration.

## 7. Troubleshooting

*   **"Cannot store password on some platforms"**: Check your system keychain permissions.
*   **"MFA verification failed"**: Ensure your MFA secret is correct and your system time is synchronized.
*   **Cross-platform compatibility issues**: `go-sshpky` abstracts platform differences using the `IKeyM` interface. If you encounter issues, please report them.

## 8. Examples

*   **Connect to a server with a specific user and port:**
    ```bash
    go-sshpky conn -u root -p 2222 my-prod-server
    ```
*   **Add a new development server to the 'dev' group:**
    ```bash
    go-sshpky add -g dev
    # Follow prompts
    ```
*   **List all hosts in the 'staging' group:**
    ```bash
    go-sshpky ms staging
    ```
