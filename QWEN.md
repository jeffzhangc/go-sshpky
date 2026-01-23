# go-sshpky Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-05

## Active Technologies

Go 1.24.0
Cobra CLI framework
Keychain integration for secure credential storage
TOTP/OTP generation for MFA
SSH connection management
YAML configuration files

## Project Structure

```text
sshpky/
├── cmd/
│   ├── conn.go
│   ├── mg.go
│   ├── ms.go
│   ├── ms_bubble.go
│   └── root.go
├── pkg/
│   ├── config/
│   ├── km/ (key manager - platform specific)
│   ├── logger/
│   ├── sshpass/
│   ├── sshrunner/
│   └── utils/
├── main.go
└── go.mod
```

## Commands

go build - Build the application
go test ./... - Run all tests
go test -cover ./... - Run tests with coverage report
make build - Build with make
sshpky [host] - Connect to a host
sshpky mg - Manage groups
sshpky ms - Manage SSH config items
sshpky conn - Connect to hosts with specific options

## Code Style

Go language standards (gofmt, golint)
Security-first approach for credential handling
Clear function and variable names
Comprehensive error handling
Documentation for exported functions (in Chinese)
Chinese comments for all code
Unit tests for all functions with 80%+ coverage

## Recent Changes

[SSH key management with secure credential storage]
[MFA/TOTP support for enhanced authentication]
[Configuration grouping and management features]

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
