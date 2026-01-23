# Data Model: SSH密钥管理功能增强

## ProjectArchitecture
**Description**: 代表项目的整体架构设计，包含模块关系、依赖结构和设计模式
**Fields**:
- name: string - 项目名称 (sshpky)
- modules: map[string]Module - 项目模块映射
- dependencies: []Dependency - 项目依赖列表
- patterns: []DesignPattern - 使用的设计模式列表

## Module
**Description**: 项目中的模块定义
**Fields**:
- name: string - 模块名称
- path: string - 模块路径 (e.g., "cmd/", "pkg/config/")
- description: string - 模块功能描述
- interfaces: []Interface - 模块中定义的接口
- responsibilities: []string - 模块职责列表

## CodeExample
**Description**: 代表特定功能的代码示例，包含实现细节和最佳实践
**Fields**:
- title: string - 示例标题
- filePath: string - 文件路径
- codeBlock: string - 代码块内容
- description: string - 代码功能说明
- bestPractices: []string - 最佳实践要点
- usageContext: string - 使用上下文

## SecurityPattern
**Description**: 代表安全相关的实现模式，包括凭证存储、加密和认证机制
**Fields**:
- name: string - 安全模式名称
- description: string - 安全模式描述
- implementation: string - 实现方式 (e.g., "Keychain", "Encrypted Storage")
- platforms: []string - 支持的平台
- validationRules: []string - 安全验证规则

## PlatformSpecific
**Description**: 代表平台特定的实现，如macOS Keychain、Linux密钥环等
**Fields**:
- platform: string - 平台名称 (darwin, linux, windows)
- implementation: string - 实现文件路径
- dependencies: []string - 平台特定依赖
- securityFeatures: []string - 安全特性

## SshConfigItem (from existing code)
**Description**: SSH配置项定义，来自现有代码的实体
**Fields**:
- Group: string - 分组名称，用于组织管理
- Host: string - 主机别名，用于ssh命令连接时使用
- HostName: string - 实际的主机地址或IP
- Port: int - SSH端口号，默认22
- User: string - 登录用户名
- IdentityFile: string - 身份认证文件路径（私钥）
- Password: string - 登录密码（通过安全方式存储）
- MFASecret: string - MFA密钥
- EditTime: string - 最后编辑时间
- ProxyCommand: string - 代理命令，用于跳板机等场景
- Desc: string - 配置项描述
- OtherParams: []string - 其他参数

## IKeyM Interface (from existing code)
**Description**: 密钥管理接口，用于跨平台凭证管理
**Methods**:
- SavePwd(connConf *SshConfigItem): 保存密码
- GetPwd(connConf SshConfigItem) string: 获取密码
- GetMAFSecret(connConf SshConfigItem) string: 获取MFA密钥

## Relationship Diagram
```
[ProjectArchitecture] 1->* [Module]
[Module] 1->* [CodeExample]
[CodeExample] -* [SecurityPattern]
[ProjectArchitecture] -* [PlatformSpecific]
[PlatformSpecific] -* [SecurityPattern]
```

## Validation Rules
- All code examples must follow Chinese documentation requirement from constitution
- Security implementations must use system keychain or equivalent secure storage
- Cross-platform code must implement the same interface contracts
- All configuration data must support serialization/deserialization
- Password/MFA data must never be stored in plain text