# Quickstart Guide: SSH密钥管理功能增强

## 项目概述
sshpky是一个安全的SSH连接管理工具，帮助开发者安全地管理多个SSH连接、凭证和环境。此功能增强通过提供详细的项目分析和代码示例，帮助开发者理解和扩展项目功能。

## 开发环境设置
1. 确保安装Go 1.24.0或更高版本
2. 克隆项目仓库
   ```bash
   git clone <repository-url>
   cd go-sshpky
   ```

3. 安装依赖
   ```bash
   go mod download
   ```

4. 构建项目
   ```bash
   go build -o sshpky
   ```

## 项目架构概览
- `cmd/` - CLI命令实现，使用Cobra框架
- `pkg/` - 核心功能包
  - `config/` - 配置管理
  - `km/` - 跨平台密钥管理
  - `sshrunner/` - SSH连接执行
  - `utils/` - 工具函数

## 关键代码示例

### 1. SSH配置项定义
```go
type SshConfigItem struct {
    Group        string // 分组名称，用于组织管理
    Host         string // 主机别名，用于 ssh 命令连接时使用
    HostName     string // 实际的主机地址或 IP
    Port         int    // SSH 端口号，默认 22
    User         string // 登录用户名
    IdentityFile string // 身份认证文件路径（私钥）
    Password     string // 登录密码（注意：明文存储密码不安全）
    MFASecret    string // mfa 密码
    EditTime     string // 最后编辑时间
    ProxyCommand string // 代理命令，用于跳板机等场景
    Desc         string // 配置项描述
    OtherParams  []string
}
```

### 2. 密钥管理接口
```go
type IKeyM interface {
    SavePwd(connConf *SshConfigItem)  // 保存密码
    GetPwd(connConf SshConfigItem) string  // 获取密码
    GetMAFSecret(connConf SshConfigItem) string  // 获取MFA密钥
}
```

### 3. 跨平台密钥链实现
```go
// Darwin implementation
func GetPassword(username, host string) (string, error) {
    return getPasswordByType(getHostKey(username, host), PWD_NORMAL)
}

func SavePassword(username, host, password string) error {
    return savePwdByType(getHostKey(username, host), PWD_NORMAL, password)
}
```

## 开发最佳实践

### 安全实践
- 密码必须通过系统密钥链存储，不得以明文保存
- 新功能必须包含安全审查
- 所有凭证传输必须使用加密方式

### 代码风格
- 所有代码注释必须使用中文
- 遵循Go语言规范
- 接口驱动设计以支持跨平台实现

### 测试要求
- 所有新功能必须包含单元测试
- 代码覆盖率必须达到80%以上
- 集成测试覆盖凭证管理流程

## 添加新功能步骤
1. 在对应的包中创建新功能代码
2. 遵循现有接口设计模式
3. 添加单元测试并确保覆盖率达标
4. 编写中文注释和文档
5. 运行所有测试确保无回归

## 调试技巧
```bash
# 构建带调试信息的版本
go build -gcflags="-N -l" -o sshpky

# 运行测试并查看覆盖率
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 常见问题
1. **无法在某些平台上存储密码**: 检查系统密钥链权限
2. **MFA验证失败**: 确认MFA密钥正确且时间同步
3. **跨平台兼容性问题**: 使用IKeyM接口抽象平台差异

## 贡献指南
- 所有PR必须通过安全、功能、测试和文档审查
- 新功能需要提供中文文档
- 遵循项目宪法要求