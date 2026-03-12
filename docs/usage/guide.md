# Go-sshpky Usage Guide

This guide provides comprehensive instructions on how to use `go-sshpky` for managing your SSH connections and credentials.

## Table of Contents

1.  [Introduction](#1-introduction)
2.  [Core Concepts](#2-core-concepts)
    *   [Configuration File](#configuration-file)
    *   [Groups](#groups)
3.  [Example Configuration Files](#3-example-configuration-files)
    *   [`~/.go-sshpky/config.yaml`](#~go-sshpkyconfigyaml)
    *   [`~/.ssh/config`](#~sshconfig)
4.  [Command: `sshpky mg` (Manage Groups)](#4-command-sshpky-mg-manage-groups)
    *   [`mg list`](#mg-list)
    *   [`mg add`](#mg-add)
    *   [`mg use`](#mg-use)
    *   [`mg delete`](#mg-delete)
5.  [Command: `sshpky ms` (Manage SSH Configs)](#5-command-sshpky-ms-manage-ssh-configs)
    *   [`ms list`](#ms-list-subcommand)
    *   [`ms` with `-l` flag](#ms-with--l-flag)
    *   [`ms add`](#ms-add)
    *   [`get`](#get)
    *   [`ms update`](#ms-update)
    *   [`ms delete`](#ms-delete)
6.  [Command: `sshpky conn` (Connect)](#6-command-sshpky-conn-connect)
    *   [Connecting to a Managed Host](#connecting-to-a-managed-host)
    *   [Direct Connection](#direct-connection)
7.  [Advanced Features](#7-advanced-features)
    *   [Multi-Factor Authentication (MFA/OTP)](#multi-factor-authentication-mfaotp)
    *   [Using Identity Files (Private Keys)](#using-identity-files-private-keys)
8.  [Example Workflow](#8-example-workflow)

---

## 1. Introduction

`go-sshpky` is a powerful command-line tool designed to simplify and secure your SSH workflow. It allows you to manage multiple SSH connections, credentials, and configurations efficiently by organizing hosts into groups and storing sensitive information in your system's keychain.

## 2. Core Concepts

### Configuration File

`go-sshpky` stores all its configuration in a YAML file located at `~/.sshpky/config.yaml`. This file is created and managed automatically. It contains your groups and host connection details. Sensitive data like passwords are not stored here directly but are handled by your system's keychain.

### Groups

Groups are the primary way to organize your SSH hosts. You can create groups for different environments (e.g., `production`, `staging`), projects, or clients. All host configurations (`ms` entries) belong to a group.

### Credential Storage Categories

`go-sshpky` supports two main ways to store sensitive credentials (like passwords or MFA secrets) associated with your SSH configurations, controlled by the `category` field in the `config.yaml`:

*   **`category: 0` (StoreFile)**: Credentials are encrypted and stored in a local file. This might be suitable for environments where a system keychain is not available or preferred.
*   **`category: 1` (StoreKeyChain)**: Credentials are stored securely in your operating system's native keychain (e.g., macOS Keychain, Windows Credential Manager, Linux Secret Service). This is generally the most secure and recommended method.

## 3. Example Configuration Files

To give you a clearer picture, here are examples of what the `go-sshpky` and standard SSH configuration files look like.

### `~/.go-sshpky/config.yaml`

This file is the main configuration for `go-sshpky` and is managed by the tool. You typically don't need to edit it by hand. Its primary role is to define your groups and their settings.

**Important:** The `secret` field contains an encrypted/encoded value that `go-sshpky` uses to manage credentials. It is **not** a plaintext password or MFA secret. You should not manually edit this field.

```yaml
# 'use' defines the currently active group.
use: work
keySize: 24
groups:
- name: work
  # The secret is an encrypted value used by go-sshpky.
  secret: "<ENCRYPTED_VALUE_1>"
  autoSave: true
  desc: Servers for my main job
  category: 0 # 0: StoreFile (local file), 1: StoreKeyChain (system keychain)
- name: personal-projects
  # A group may not have a secret.
  secret: ""
  autoSave: true
  desc: Hobby project servers and VMs
  category: 1 # 0: StoreFile (local file), 1: StoreKeyChain (system keychain)
- name: testing-lab
  secret: "<ENCRYPTED_VALUE_2>"
  autoSave: true
  desc: Temporary cloud instances for testing
  category: 0 # 0: StoreFile (local file), 1: StoreKeyChain (system keychain)
```

*Note: The host connection details (like IP addresses, users, etc.) are managed separately by the `sshpky ms` command and are stored in a different configuration structure that is linked to these groups.*


### `~/.ssh/config`

While `go-sshpky` does not manage this file, it's good practice to have a global SSH config for settings that affect all connections, including those made by `go-sshpky`.

```
# This file is for global SSH client settings.
# It is NOT managed by go-sshpky, but its settings can affect go-sshpky's connections.

# Enable connection sharing to speed up subsequent connections to the same host.
Host *
    ControlMaster auto
    ControlPath ~/.ssh/master-%r@%h:%p
    ControlPersist 10m

# Keep connections alive by sending a packet every 60 seconds.
    ServerAliveInterval 60
    ServerAliveCountMax 3
```

### `~/.ssh/config` Entry Generated by `go-sshpky`

`go-sshpky` might generate and add entries like the following to your `~/.ssh/config` file when you connect to a host. These entries often contain encrypted credentials in comments, which `go-sshpky` uses for authentication.

```
# Group: work
# EditTime: 2026-03-12 15:00:00
# Password: <ENCRYPTED_PASSWORD_STRING>
# MFASecret: <ENCRYPTED_MFA_SECRET_STRING>
Host example-host-alias
    HostName 192.168.1.100
    Port 22
    User webadmin
```

## 4. Command: `sshpky mg` (Manage Groups)

This command is used to manage your configuration groups.

### `mg list`

Lists all the groups you have created.

```bash
sshpky mg list
```

### `mg add`

Adds a new, empty group.

```bash
# Add a group named 'production'
sshpky mg add production
```

### `mg use`

Sets a group as the default for all subsequent commands. When you run `sshpky ms add`, the new host will be added to this default group.

```bash
# Set 'production' as the default group
sshpky mg use production
```

### `mg delete`

Deletes a group and all the host configurations within it.

```bash
sshpky mg delete production
```

## 5. Command: `sshpky ms` (Manage SSH Configs)

This command manages the individual SSH host configurations within your groups. It can be run interactively or in a non-interactive list mode.

- **Interactive Mode (default)**: Running `sshpky ms` without flags starts a full-featured terminal UI where you can browse, search, add, update, and delete hosts.
- **Non-Interactive Mode (`-l` or `--list`)**: Use the `-l` flag to quickly list hosts without entering the interactive UI.

You can specify a group with the `-g` flag; otherwise, it operates on the current default group (set by `sshpky mg use`).

### `ms list` (Subcommand)

Lists all SSH host configurations. This is similar to `sshpky ms -l`.

```bash
# List hosts in the default group
sshpky ms list

# List hosts in the 'staging' group
sshpky ms list staging
```

### `ms` with `-l` flag

Provides a quick, non-interactive list of hosts.

```bash
# List hosts in the default group
sshpky ms -l

# List hosts in the 'staging' group
sshpky ms -l -g staging
```

### `ms add`

Interactively adds a new SSH host configuration to a group.

```bash
# Add a host to the default group
sshpky ms add
```

The tool will prompt you for the host alias, connection details (`user@host`), and credentials.

### `get`

Retrieves and displays the details of a specific host configuration.

```bash
# Get details for a host in the default group
sshpky get <host-alias>

# Get details for a host in a specific group
sshpky get <host-alias> -g <group-name>

# Get the configuration in YAML format
sshpky get <host-alias> -o yaml

# Get only the password for a host
sshpky get <host-alias> -k password
```

### `ms update`

Allows you to modify an existing SSH host configuration.

```bash
sshpky ms update <host-alias>
```

### `ms delete`

Deletes a specific SSH host configuration.

```bash
sshpky ms delete <host-alias>
```

## 6. Command: `sshpky conn` (Connect)

This is the command used to initiate an SSH connection.

### Connecting to a Managed Host

You can connect to any host configured with `sshpky ms` by using its alias. `go-sshpky` will find its configuration and connect automatically.

```bash
sshpky conn <host-alias>
```

The top-level `sshpky` command is a shortcut for this:
```bash
sshpky <host-alias>
```

### Direct Connection

You can also use `conn` to connect to a host that is not in your configuration, similar to the standard `ssh` command.

```bash
sshpky conn user@example.com
sshpky conn -p 2222 -i ~/.ssh/my_key user@example.com
```

## 7. Advanced Features

### Multi-Factor Authentication (MFA/OTP)

When adding or updating a host with `sshpky ms`, you can provide a TOTP secret. When you connect, `go-sshpky` will automatically calculate the current one-time password and use it for authentication.

### Using Identity Files (Private Keys)

During the `sshpky ms add` process, you can choose "private key" as the authentication method and provide the path to your identity file (e.g., `~/.ssh/id_rsa`).

## 8. Example Workflow

Here is a complete example of setting up a new project:

```bash
# 1. Create a group for the new project
sshpky mg add my-project

# 2. Set it as the default group
sshpky mg use my-project

# 3. Add the project's web server
# Follow the interactive prompts
sshpky ms add

# Enter 'web-server' as the alias
# Enter 'deploy@192.0.2.10' as the user@host
# ...and provide credentials

# 4. Add the project's database server
sshpky ms add

# Enter 'db-server' as the alias
# Enter 'admin@192.0.2.11' as the user@host
# ...and provide credentials

# 5. List the hosts in your project group
sshpky ms list

# 6. Connect to the web server
sshpky conn web-server
```
