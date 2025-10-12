# sshpky

ssh client wrapper

替换 https://github.com/jeffzhangc/sshpky_zsh_plugin

## feature

1. 记录密码
2. 记录 google auth secret
3. 自动使用上传成功的密码登录
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
    RequestTTY force
    RemoteCommand env TERM=xterm-256color bash --login
```

## referece

1. [tssh](https://github.com/luanruisong/tssh)
2. [trzsz-ssh](https://github.com/trzsz/trzsz-ssh)
3. [crypto](https://github.com/golang/crypto)
4. [goexpect](https://github.com/google/goexpect)
5. [sshpass](https://github.com/billcoding/sshpass)
6. [otp](https://github.com/pquerna/otp)
