# API Contracts: SSH密钥管理功能接口

## IKeyM Interface Contract
**Purpose**: 跨平台密钥管理接口，用于安全存储和检索SSH凭证

### SavePwd Method
- **Signature**: `SavePwd(connConf *SshConfigItem)`
- **Purpose**: 保存SSH连接的密码到安全存储
- **Parameters**: 
  - connConf: SshConfigItem指针，包含连接配置信息
- **Side Effects**: 将密码加密存储到系统密钥链
- **Error Handling**: 返回错误如果存储失败
- **Security**: 密码必须以加密形式存储，不得明文保存

### GetPwd Method
- **Signature**: `GetPwd(connConf SshConfigItem) string`
- **Purpose**: 从安全存储中获取SSH连接的密码
- **Parameters**: 
  - connConf: SshConfigItem，包含连接配置信息
- **Return**: 密码字符串，如果未找到则返回空字符串
- **Security**: 访问需要适当的系统权限

### GetMAFSecret Method
- **Signature**: `GetMAFSecret(connConf SshConfigItem) string`
- **Purpose**: 从安全存储中获取SSH连接的MFA密钥
- **Parameters**: 
  - connConf: SshConfigItem，包含连接配置信息
- **Return**: MFA密钥字符串，如果未找到则返回空字符串
- **Security**: MFA密钥必须以加密形式存储

## SshConfigItem Structure Contract
**Purpose**: SSH连接配置项的数据结构定义

### Fields
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

## SSH Connection Contract
**Purpose**: SSH连接操作的接口定义

### Connect Method
- **Purpose**: 建立SSH连接
- **Parameters**: SshConfigItem，包含连接所需的所有配置
- **Return**: 连接句柄或错误
- **Behavior**: 
  - 自动使用存储的凭证进行身份验证
  - 支持密码认证、密钥认证、MFA认证
  - 自动处理连接过程中的交互

## Configuration Management Contract
**Purpose**: 配置管理操作的接口定义

### Save/Load Configuration
- **Purpose**: 保存和加载SSH配置
- **Format**: YAML格式
- **Security**: 密码和MFA密钥不直接存储在配置文件中
- **Location**: 用户配置目录 (e.g., ~/.sshpky/)