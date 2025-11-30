# Development Guide

This document contains information for developers working on pvec.

## Architecture

The project follows clean architecture principles with clear separation of concerns:

```
pvec/
├── cmd/pvec/          # Main application entry point
├── pkg/
│   ├── actions/       # Action interfaces and implementations
│   ├── config/        # Configuration management
│   ├── models/        # Data models (VMStatus, NodeList)
│   ├── proxmox/       # Proxmox API client
│   └── ui/            # Bubble Tea TUI components
│       ├── mainlist/      # Main interactive list
│       ├── helpdialog/    # Help text generator
│       ├── configpanel/   # Config editor (Bubble Tea model)
│       ├── actiondialog/  # Action progress dialogs
│       └── detailsdialog/ # VM/CT details display
├── examples/
│   └── test-client/   # CLI test client
├── scripts/           # Code analysis tools
└── docs/              # Documentation
```

## Key Design Patterns

- **Bubble Tea Architecture**: Model-View-Update (MVU) pattern for UI components
- **Dependency Injection**: All components receive dependencies via constructors
- **Interface-based Design**: Client, Executor, DataProvider, Loader interfaces
- **Clean Separation**: UI components don't directly interact with API client
- **Testability**: Comprehensive test coverage for business logic with mock implementations
- **Hybrid UI Pattern**: Full Bubble Tea models for stateful components, stateless generators for simple displays

## Prerequisites

- Go 1.23 or later
- Make (optional, for convenience commands)
- golangci-lint (for linting)
- Python 3 (for code analysis)

## Building

```bash
# Build main application
make build

# Build test client
make test-client

# Clean build artifacts
make clean

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Full pipeline (clean, fmt, lint, test, build)
make
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with race detector
make test

# Run with verbose output
make test-verbose

# Generate coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage.out
```

For detailed test coverage metrics by package, see [Code Analysis Report](code_analysis.md).

### Writing Tests

- Use table-driven tests where appropriate
- Mock external dependencies (Proxmox client, etc.)
- Test both success and error paths
- Use `testify/assert` for assertions
- Maintain test coverage above 80%

## Code Quality

### Code Analysis

Run comprehensive code analysis:

```bash
# Analyze code quality and metrics
make analyze

# View detailed report
cat docs/code_analysis.md
```

The analysis includes:
- Cyclomatic complexity analysis
- Cognitive complexity analysis  
- Static analysis (go vet, staticcheck)
- Security analysis (gosec)
- Vulnerability scanning (govulncheck)
- Code smells detection
- Dead code detection
- Architecture validation
- Test coverage metrics
- Code quality scores

For current metrics, complexity guidelines, and detailed analysis results, see [Code Analysis Report](code_analysis.md).

### Code Standards

- Follow Go best practices and idioms
- Use `gofmt` for formatting (run `make fmt`)
- Pass all linter checks (`make lint`)
- Add comments for exported functions/types
- Keep functions small and focused
- Extract complex logic into helper functions

## Examples

### Test Client (CLI)

The `test-client` example demonstrates API usage without TUI:

```bash
# Build and run test client
go build -o bin/test-client examples/test-client/main.go
./bin/test-client -c test-config.json

# Or run directly
go run examples/test-client/main.go -c test-config.json
```

Displays nodes and their VMs/containers organized by Proxmox node, showing:
- Node-grouped VM/CT listing
- Status, type, resource usage
- Connection validation

**Features:**
- Simple terminal output (no TUI)
- Configuration validation
- API connectivity testing
- Organized display by Proxmox nodes

### Test UI (Interactive)

The `test-ui` example shows the main list component:

```bash
go run examples/test-ui/main.go -c test-config.json
```

Interactive TUI with keyboard navigation (without action execution).

## UI Development

### Bubble Tea Components

The UI uses two patterns:

1. **Full Bubble Tea Models** (for stateful, interactive components):
   - Implement `Init() tea.Cmd`, `Update(tea.Msg) (tea.Model, tea.Cmd)`, `View() string`
   - Maintain internal state (input fields, cursor positions, etc.)
   - Example: `configpanel.Model`
   - Must receive `tea.WindowSizeMsg` and capture returned model during initialization

2. **Stateless Text Generators** (for simple displays):
   - Functions that take width/height parameters and return formatted strings
   - No state to maintain, rendered fresh each call
   - Example: `helpdialog.GetHelpText(width, height)`
   - Simpler and faster for non-interactive content

### Adding New UI Components

1. Create new package under `pkg/ui/`
2. Choose pattern: Full Bubble Tea model (stateful) or text generator (stateless)
3. Accept dependencies via constructor (no globals)
4. For Bubble Tea models: Implement Init/Update/View, handle WindowSizeMsg
5. Create interfaces for testability
6. Document exported functions

## Contributing

### Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Write tests for new functionality
5. Ensure all checks pass:
   ```bash
   make fmt
   make lint
   make test
   make build
   ```
6. Commit changes (`git commit -am 'Add amazing feature'`)
7. Push to branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Commit Messages

- Use clear, descriptive commit messages
- Start with a verb (Add, Fix, Update, Refactor, etc.)
- Reference issues when applicable (#123)

### Pull Request Guidelines

- Provide clear description of changes
- Link related issues
- Ensure CI passes
- Update documentation if needed
- Add tests for new features
- Maintain or improve code coverage

## Debugging

### Running with Debug Output

```bash
# Build with debug symbols
go build -o ./bin/pvec .

# Run with verbose logging (if implemented)
./bin/pvec -c config.json
```

### Common Issues

**UI not rendering correctly**:
- Check terminal size (minimum 80x24 recommended)
- Verify terminal supports 256 colors
- Try different terminal emulators

**Test failures**:
- Ensure no stale mocks
- Check race conditions with `-race` flag
- Verify test data matches current types

**Build failures**:
- Run `go mod tidy`
- Check Go version (1.23+ required)
- Clear build cache: `go clean -cache`

## Release Process

1. Update version in `main.go`
2. Update CHANGELOG (if exists)
3. Run full test suite: `make`
4. Tag release: `git tag v1.x.x`
5. Push tags: `git push --tags`
6. Build release binaries
7. Create GitHub release with binaries

## Tools and Dependencies

### Runtime Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) v1.3.10 - Terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components (textinput, etc.)
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [viper](https://github.com/spf13/viper) - Configuration management

### Development Dependencies

- [testify](https://github.com/stretchr/testify) - Testing toolkit
- [golangci-lint](https://golangci-lint.run/) - Linters aggregator
- [gosec](https://github.com/securego/gosec) - Security checker
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) - Vulnerability scanner

### Installing Development Tools

```bash
# Install golangci-lint
brew install golangci-lint  # macOS
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

## Additional Resources

- [Code Analysis Report](code_analysis.md) - Detailed code metrics
- [API Documentation](https://pkg.go.dev/github.com/tsupplis/pvec) - Go package docs
- [Proxmox API Reference](https://pve.proxmox.com/pve-docs/api-viewer/) - Proxmox API docs
- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials) - Official tutorials
