# sshpky

ssh client wrapper

替换 https://github.com/jeffzhangc/sshpky_zsh_plugin

## feature

1. 记录密码
2. 记录 google auth secret
3. 自动使用上次成功登陆的密码登录
4. 如果有 otp 验证，自动使用上次记录的 secret 生成 otp 验证码进行尝试登录
5. 新增 保存 连接成功的 ssh 信息 在配置文件中
6. 删除 配置信息
7. 配置信息分组

## install

```
brew install jeffzhangc/tap/sshpky

```

## build

```
make build
make install
```

## TODO

完善文档和使用图片

```
sshpky -h
A comprehensive tool for managing SSH public keys and connections.

Simplify your SSH key management and server connections with automatic
key selection, group management, and secure credential storage.

If a HOST argument is provided, it will automatically connect to that host
using the connection command. Otherwise, you can use subcommands for
specific operations.

Examples:
  # Quick connect to a host (default behavior)
  sshpky 10.1.102.32
  sshpky user@example.com

  # Connect with specific options
  sshpky -u admin -p 2222 example.com
  sshpky -g production webserver

  # Use subcommands for other operations
  sshpky mg list                    # List all groups
  sshpky mg use production         # Switch to production group
  sshpky conn -g staging appserver  # Connect to specific group's host
  sshpky ms 	# manage ssh configItem for current group
  sshpky ms groupA	# manage ssh configItem for gropuA

Usage:
  sshpky [flags] [HOST]
  sshpky [command]

Available Commands:
  completion  Generate shell completion scripts
  conn        Connet to a host using managed SSH keys
  help        Help about any command
  mg          Manage SSH key groups
  ms          Manage SSH config items

Flags:
  -f, --config string       config file name (default "config.yaml")
  -c, --config-dir string   config directory (default "/Users/xxx/.sshpky")
  -g, --group string        group for this ssh
  -h, --help                help for sshpky
      --hostname string     hostname for this ssh
  -i, --identity string     identity file (private key) for public key authentication
  -p, --port int            SSH port (default 22)
  -u, --user string         username for SSH connection

Use "sshpky [command] --help" for more information about a command.
```

## ssh config

common config

```
ControlMaster auto
ControlPath ~/.ssh/master-%r@%h:%p
Host *
    HostKeyAlgorithms +ssh-rsa
    PubkeyAcceptedKeyTypes +ssh-rsa
    ServerAliveInterval 60
    ServerAliveCountMax 3
```

## referece

1. [tssh](https://github.com/luanruisong/tssh)
2. [trzsz-ssh](https://github.com/trzsz/trzsz-ssh)
3. [crypto](https://github.com/golang/crypto)
4. [goexpect](https://github.com/google/goexpect)
5. [sshpass](https://github.com/billcoding/sshpass)
6. [otp](https://github.com/pquerna/otp)
