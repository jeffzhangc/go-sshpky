# Research: SSH密钥管理功能增强

## Decision: Project Architecture Analysis
**Rationale**: To enhance the SSH management tool, a comprehensive understanding of the existing architecture is essential to maintain consistency and follow established patterns.

## Project Structure Analysis
- **CLI Layer** (`cmd/`): Uses Cobra framework for command-line interface
- **Core Logic** (`pkg/`): Modular design with specific packages for different concerns
- **Configuration Management** (`pkg/config/`): Handles SSH configurations, groups, and credential storage
- **Key Management** (`pkg/km/`): Platform-specific credential storage (macOS Keychain, Linux secret service, Windows)
- **SSH Operations** (`pkg/sshrunner/`): Handles actual SSH connections and terminal interactions
- **Utilities** (`pkg/utils/`): Helper functions including TOTP generation for 2FA

## Technology Stack Analysis
- **Language**: Go 1.24.0
- **CLI Framework**: Cobra for command-line parsing and help generation
- **Platform Integration**: keybase/go-keychain for secure credential storage
- **SSH Handling**: golang.org/x/crypto and custom implementations
- **Terminal Interaction**: google/goexpect for automated SSH sessions
- **UI Components**: charmbracelet/bubbletea for TUI elements
- **Authentication**: pquerna/otp for TOTP/2FA code generation

## Design Patterns Identified
1. **Interface-based Architecture**: IKeyM interface for cross-platform key management
2. **Configuration-Driven**: YAML files for SSH settings and user preferences
3. **Command Pattern**: Cobra commands for CLI operations
4. **Strategy Pattern**: Different key storage strategies for different platforms
5. **Separation of Concerns**: Clear division between CLI, business logic, and platform-specific code

## Security Considerations
- Credentials stored in system keychain, not plain text
- Passwords encrypted using OS-provided secure storage
- MFA/2FA support through TOTP generation
- Secure credential transmission during SSH connections
- Group-based configuration isolation

## Cross-Platform Implementation
- **macOS**: Uses Keychain API via keybase/go-keychain
- **Linux**: Uses secret service API
- **Windows**: Implementation for password storage
- Consistent interface across platforms through IKeyM

## Testing Strategy
- Unit tests with 80%+ coverage requirement (per constitution)
- Integration tests for SSH connection workflows
- Platform-specific tests for keychain implementations
- CLI behavior tests for command flows

## CLI Architecture
- Primary command: `sshpky [host]` for direct connection
- Subcommands: `mg` (manage groups), `ms` (manage SSH configs), `conn` (connect with options)
- Global flags for configuration directory, user, port, etc.
- Tab completion for hosts and groups
- Human-friendly and JSON output options

## Alternatives Considered
- Using different configuration formats (JSON, TOML): Chose YAML for readability
- Different CLI frameworks: Cobra provides the best feature set for this use case
- Alternative secure storage: System keychain is the most secure approach
- Different SSH libraries: Go's crypto libraries provide the necessary functionality