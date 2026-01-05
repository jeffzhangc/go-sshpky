# 功能规格说明：SSH密钥管理功能增强（添加项目分析和示例代码）

**功能分支**: `002-enhance-ssh-mgmt`  
**创建日期**: 2026-01-05  
**状态**: 草稿  
**输入**: 用户描述: "帮我完善 001-secure-ssh-mgmt 添加 分析当前项目,示例代码到 规范里,为后续生成代码时做参考"

## 用户场景与测试 *(必填)*

### 用户故事 1 - 项目分析集成 (优先级: P1)

作为开发者，我希望能够看到当前项目(go-sshpky)的完整分析，包括架构、模块设计、依赖关系和代码结构，这样我可以在开发新功能时参考现有实现模式。

**为什么此优先级**: 了解现有项目架构是确保新功能与现有代码保持一致性的关键前提。

**独立测试**: 可以通过查看项目分析文档来验证项目结构和设计模式是否被正确记录和理解。

**验收场景**:

1. **Given** 项目分析文档已完成, **When** 开发者查阅文档, **Then** 应能理解项目的整体架构和模块关系
2. **Given** 开发者需要参考现有实现, **When** 查看示例代码, **Then** 应能找到类似的实现模式作为参考

---

### 用户故事 2 - 示例代码参考 (优先级: P2)

作为开发者，我希望能够访问现有代码的示例和最佳实践，以便在实现新功能时保持代码风格和架构一致性。

**为什么此优先级**: 示例代码是确保代码质量、一致性和最佳实践的关键资源。

**独立测试**: 可以通过查阅提供的示例代码来验证开发者能否基于现有模式编写新功能。

**验收场景**:

1. **Given** 需要实现新的SSH配置功能, **When** 开发者参考示例代码, **Then** 应能按照现有模式实现功能
2. **Given** 开发者需要添加安全认证功能, **When** 查看安全相关示例, **Then** 应能正确集成安全机制

---

### 用户故事 3 - 代码生成参考 (优先级: P3)

作为开发者，我希望规格说明包含完整的项目上下文信息，以便为后续的代码生成工具提供充分的上下文支持。

**为什么此优先级**: 完整的上下文信息是确保代码生成质量的关键因素。

**独立测试**: 可以通过将项目分析信息提供给代码生成工具来验证它是否能生成符合项目架构的代码。

**验收场景**:

1. **Given** 代码生成工具需要项目上下文, **When** 提供项目分析信息, **Then** 应能生成符合架构的代码
2. **Given** 需要生成安全相关功能, **When** 参考安全实现模式, **Then** 应能正确实现安全功能

---

### 边界情况

- 当项目结构发生变化时，如何确保分析信息保持同步？
- 如何处理不同平台（macOS、Linux、Windows）的实现差异？
- 当现有代码模式不适用于新需求时，如何记录和处理？

## 项目分析与代码示例参考

### 项目架构分析

#### 主要模块结构：
- `cmd/` - CLI命令实现（root.go, conn.go, mg.go, ms.go, ms_bubble.go）
- `pkg/` - 核心功能包
  - `config/` - 配置管理（config.go, groupm.go, keychainM.go, keyStoreFileM.go, sshconfigm.go）
  - `km/` - 密钥管理（平台特定实现：darwin, linux, windows）
  - `logger/` - 日志记录
  - `sshpass/` - SSH密码处理
  - `sshrunner/` - SSH运行器
  - `utils/` - 工具函数（totp.go, secret.go, utils.go）

#### CLI架构：
- 使用Cobra框架实现命令行接口
- 支持`sshpky [host]`直接连接模式和子命令模式
- 支持多种参数（用户、端口、密钥文件、分组等）
- 具备自动补全功能

#### 安全架构：
- 密码通过系统密钥链存储（macOS Keychain, Linux secret service等）
- 支持TOTP/MFA认证
- 配置按分组管理，支持不同安全策略

### 关键代码示例

#### 配置项定义示例（pkg/config/config.go）：
```
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

#### 密钥管理接口示例（pkg/config/config.go）：
```
type IKeyM interface {
    SavePwd(connConf *SshConfigItem)
    GetPwd(connConf SshConfigItem) string
    GetMAFSecret(connConf SshConfigItem) string
}
```

#### 跨平台密钥链实现示例（pkg/km/keymanager_darwin.go）：
```
func GetPassword(username, host string) (string, error) {
    return getPasswordByType(getHostKey(username, host), PWD_NORMAL)
}

func SavePassword(username, host, password string) error {
    return savePwdByType(getHostKey(username, host), PWD_NORMAL, password)
}
```

#### SSH连接实现示例（pkg/sshrunner/sshrunner.go）：
```
// SSH连接主要通过PTY（伪终端）实现
// 支持自动输入密码、MFA码等
// 使用goexpect库处理交互式终端
```

### 设计模式与最佳实践

1. **接口驱动设计**：通过IKeyM接口实现跨平台密钥管理
2. **配置驱动**：通过YAML配置文件管理SSH连接设置
3. **分组管理**：通过分组概念组织不同环境的连接
4. **安全优先**：密码通过系统密钥链存储，避免明文存储
5. **CLI优先**：所有功能都通过命令行接口提供

## 要求 *(必填)*

### 功能要求

- **FR-001**: 系统必须提供完整的项目架构分析
- **FR-002**: 系统必须包含关键模块的代码示例
- **FR-003**: 系统必须记录安全实现的最佳实践
- **FR-004**: 系统必须提供CLI接口实现的参考模式
- **FR-005**: 系统必须记录单元测试的实现模式
- **FR-006**: 系统必须包含配置管理的实现参考
- **FR-007**: 系统必须提供跨平台实现的差异说明
- **FR-008**: 系统必须记录依赖管理和版本控制的实践
- **FR-009**: 系统必须提供错误处理和日志记录的参考实现
- **FR-010**: 系统必须包含性能优化和安全考虑的实践
- **FR-011**: 系统必须提供项目现有的SSH连接实现模式
- **FR-012**: 系统必须包含项目现有的配置管理实现模式
- **FR-013**: 系统必须提供项目现有的MFA/OTP实现参考
- **FR-014**: 系统必须提供项目现有的用户界面实现模式（BubbleTea）
- **FR-015**: 系统必须描述项目现有的测试策略和覆盖率要求

### 关键实体

- **ProjectArchitecture**: 代表项目的整体架构设计，包含模块关系、依赖结构和设计模式
- **CodeExample**: 代表特定功能的代码示例，包含实现细节和最佳实践
- **SecurityPattern**: 代表安全相关的实现模式，包括凭证存储、加密和认证机制
- **PlatformSpecific**: 代表平台特定的实现，如macOS Keychain、Linux密钥环等

## 成功标准 *(必填)*

### 可测量结果

- **SC-001**: 开发者能在10分钟内通过项目分析了解核心架构
- **SC-002**: 示例代码覆盖项目80%的核心功能模块
- **SC-003**: 90%的新功能开发能够参考现有模式实现
- **SC-004**: 代码生成工具能够基于项目分析生成70%的初始代码
- **SC-005**: 文档化的新功能开发成功率提高至95%