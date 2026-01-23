<!-- 
Sync Impact Report:
Version change: 1.1.0 → 1.2.0
Added sections: Chinese Documentation Principle (Section VI)
Modified sections: Core Principles (added Chinese documentation requirement)
Templates requiring updates: .specify/templates/* (need to update references to Chinese documentation)
Follow-up TODOs: Update template files to reflect new Chinese documentation requirements

File Location: This constitution is now stored in the docs/governance/ directory
as part of the new documentation structure.
-->
# go-sshpky 项目宪法

## 核心原则

### I. 安全优先
所有功能必须优先考虑安全性；密码和敏感信息必须加密存储；支持密钥链集成以提供安全的凭证管理；任何可能暴露敏感信息的功能必须经过安全审查。

### II. CLI 界面
每个功能都应通过命令行接口提供；支持标准输入/输出协议：stdin/args → stdout，错误 → stderr；提供 JSON 和人性化格式输出。

### III. 测试优先（不可协商）
所有功能必须先编写测试；遵循 TDD 周期：编写测试 → 用户批准 → 测试失败 → 实现功能 → 重构；严格遵循红-绿-重构周期；代码覆盖率必须达到80%以上。

### IV. 单元测试要求
所有函数、方法和包必须有对应的单元测试；任何新功能或修复必须包含相应的单元测试；单元测试必须覆盖所有代码路径，包括边界条件和错误处理。

### V. 集成测试
重点关注需要集成测试的领域：新功能的端到端测试、配置更改、跨模块通信、安全凭证管理流程。

### VI. 中文文档与注释
所有代码注释、技术文档和用户说明必须使用中文编写；确保中文为项目的首要交流语言；提高中文开发者的可读性和维护性。

### VII. 可观测性与版本控制
所有操作必须提供结构化日志记录；实现调试友好性；遵循 MAJOR.MINOR.BUILD 版本格式；确保向后兼容性。

## 安全要求

密码和MFA密钥管理：支持安全的密码和TOTP密钥存储，使用系统密钥链（如macOS Keychain）或安全的文件存储；密码必须加密存储，不得以明文形式保存。

SSH连接安全：支持SSH密钥认证、密码认证和MFA认证；自动使用上次成功登录的凭据进行连接；提供安全的代理命令支持。

配置管理：支持按组管理SSH配置；配置文件必须以安全方式存储；支持配置项的增删改查操作。

## 开发工作流程

代码审查：所有PR必须经过安全、功能、测试和文档审查；复杂功能必须提供中文文档；遵循Go语言最佳实践。

质量门控：所有单元测试和集成测试必须通过；代码覆盖率必须达到80%以上；代码必须符合Go语言规范；安全功能必须经过专门审查；所有代码注释和文档必须使用中文。

测试要求：每个PR必须包含充分的单元测试；新功能必须有对应的测试用例；修复的bug必须有回归测试；所有测试必须在CI环境中通过。

文档要求：所有代码注释必须使用中文；所有技术文档必须使用中文；所有用户说明必须使用中文；提高中文开发者的可读性和可维护性。

部署流程：使用标准Go构建工具；支持通过包管理器安装（如Homebrew）；提供跨平台支持（Linux、macOS、Windows）。

## 治理
本宪法优于所有其他实践；修订需要文档记录、批准和迁移计划；所有PR/审查必须验证合规性；复杂性必须被证明是合理的。

**Version**: 1.2.0 | **Ratified**: 2026-01-05 | **Last Amended**: 2026-01-05