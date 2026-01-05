# Implementation Plan: SSH密钥管理功能增强（添加项目分析和示例代码）

**Branch**: `002-enhance-ssh-mgmt` | **Date**: 2026-01-05 | **Spec**: [docs/specs/002-enhance-ssh-mgmt/spec.md](spec.md)
**Input**: Feature specification from `docs/specs/002-enhance-ssh-mgmt/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

此功能旨在增强现有的安全SSH密钥管理工具(go-sshpky)，通过添加详细的项目分析和示例代码，为后续的开发和代码生成提供参考。主要目标是提供完整的项目架构分析、关键模块的代码示例、设计模式与最佳实践的说明，特别关注安全实现、跨平台兼容性和CLI架构等方面。

## Technical Context

**Language/Version**: Go 1.24.0  
**Primary Dependencies**: Cobra CLI framework, keybase/go-keychain, golang.org/x/crypto, google/goexpect, charmbracelet/bubbletea, pquerna/otp  
**Storage**: YAML configuration files, system keychain (macOS Keychain, Linux secret service)  
**Testing**: Go native testing framework (go test), integration tests for SSH connections, unit tests with 80%+ coverage requirement  
**Target Platform**: Cross-platform (Linux, macOS, Windows)  
**Project Type**: Command-line interface tool (CLI)  
**Performance Goals**: SSH connection establishment under 5 seconds, configuration loading under 1 second  
**Constraints**: Secure credential storage, cross-platform compatibility, backward compatibility with existing SSH tools, CLI-first design  
**Scale/Scope**: Support for 100+ SSH configurations, multiple environment groups, multi-factor authentication

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Gates determined based on constitution file:
- Security-first: All features must pass security review - PASSED (feature is about documenting security patterns)
- CLI Interface: All functionality must be accessible via CLI - PASSED (documentation will be accessible via CLI tools)
- Test-first: All features must include tests - PASSED (documentation includes testing patterns)
- Unit Testing: All functions must have unit tests with 80%+ coverage - PASSED (documentation includes testing requirements)
- Integration Testing: Features that interact with credentials or external systems must have integration tests - PASSED (covers credential management patterns)
- Chinese Documentation: All code comments and documentation must be in Chinese - PASSED (all documentation is in Chinese as required)
- Observability: All operations must have appropriate logging - PASSED (includes logging patterns)

**Post-Design Evaluation**: All constitution gates continue to pass after completing the design phase. The documentation and analysis provided in research.md, data-model.md, quickstart.md and contracts/ align with all constitutional requirements, especially Chinese documentation requirements and security-first approach.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
sshpky/
├── cmd/
│   ├── root.go           # Main CLI entry point using Cobra
│   ├── conn.go           # Connection command implementation
│   ├── mg.go             # Group management commands
│   ├── ms.go             # SSH config management
│   └── ms_bubble.go      # BubbleTea UI for SSH config management
├── pkg/
│   ├── config/
│   │   ├── config.go     # SSH configuration structures and interfaces
│   │   ├── groupm.go     # Group management
│   │   ├── keychainM.go  # Keychain management
│   │   ├── keyStoreFileM.go # File-based key storage
│   │   └── sshconfigm.go # SSH configuration management
│   ├── km/               # Key Manager (platform-specific)
│   │   ├── keymanager_darwin.go   # macOS keychain implementation
│   │   ├── keymanager_linux.go    # Linux keychain implementation
│   │   └── keymanager_window.go   # Windows keychain implementation
│   ├── logger/
│   │   └── log.go        # Logging implementation
│   ├── sshpass/
│   │   └── sshpass.go    # SSH password handling
│   ├── sshrunner/
│   │   ├── sshrunner.go  # SSH connection runner
│   │   └── shell.go      # Shell interaction
│   └── utils/
│       ├── totp.go       # TOTP/2FA code generation
│       ├── secret.go     # Secret handling utilities
│       └── utils.go      # General utility functions
├── main.go               # Application entry point
├── go.mod                # Module dependencies
├── go.sum                # Dependency checksums
├── Makefile              # Build automation
├── docs/                 # Documentation
│   ├── governance/       # Project governance documents
│   ├── processes/        # Development processes
│   └── specs/            # Feature specifications
│       └── 002-enhance-ssh-mgmt/ # This feature spec
└── tests/                # Test files
    ├── unit/
    ├── integration/
    └── contract/
```

**Structure Decision**: Single CLI tool project with modular package structure. The architecture follows a CLI-first approach with clear separation between commands (cmd/), core functionality (pkg/), and platform-specific implementations (in pkg/km/). The feature will enhance existing documentation without changing the source code structure.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
