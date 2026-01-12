# Contributing to macOS Supply-Chain Monitor

First off, thanks for taking the time to contribute! 🎉

The following is a set of guidelines for contributing to macOS Supply-Chain Monitor. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## Code of Conduct

This project and everyone participating in it is governed by our commitment to creating a welcoming and inclusive environment. Be respectful, constructive, and professional.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (command outputs, logs, etc.)
- **Describe the behavior you observed** and what you expected
- **Include your environment details**:
  - macOS version
  - Go version
  - scm version (`scm --version`)
  - Package managers installed

**Bug Report Template:**
```markdown
### Description
[Clear description of the bug]

### Steps to Reproduce
1. Run `scm daemon start`
2. Install package with `brew install ...`
3. ...

### Expected Behavior
[What you expected to happen]

### Actual Behavior
[What actually happened]

### Environment
- macOS: 14.2
- Go: 1.21.5
- scm: 0.1.0
- Package Managers: Homebrew 4.2.0, npm 10.2.0

### Logs
```
[Paste relevant log output]
```
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use a clear and descriptive title**
- **Provide a step-by-step description** of the suggested enhancement
- **Provide specific examples** to demonstrate the steps
- **Describe the current behavior** and **explain the behavior you'd like to see**
- **Explain why this enhancement would be useful**

### Pull Requests

The process described here has several goals:

- Maintain code quality
- Fix problems that are important to users
- Engage the community in working toward the best possible tool
- Enable a sustainable system for maintainers to review contributions

Please follow these steps:

1. **Fork the repo** and create your branch from `main`
2. **Make your changes**
   - If you've added code that should be tested, add tests
   - Ensure the test suite passes (`make test`)
   - Make sure your code follows the existing style (`make lint`)
3. **Commit your changes** using clear commit messages
4. **Push to your fork** and submit a pull request

#### Pull Request Guidelines

- **Keep PRs focused** - One feature/fix per PR
- **Write clear commit messages** - Use present tense ("Add feature" not "Added feature")
- **Update documentation** - If you change APIs or behavior
- **Add tests** - For new features or bug fixes
- **Follow Go conventions** - Use `gofmt`, follow standard library patterns

**PR Template:**
```markdown
### Description
[Brief description of what this PR does]

### Motivation
[Why is this change necessary? What problem does it solve?]

### Changes
- [List of changes]
- [Another change]

### Testing
- [ ] Unit tests added/updated
- [ ] Manual testing performed
- [ ] Documentation updated

### Checklist
- [ ] Code follows project style
- [ ] Tests pass locally (`make test`)
- [ ] Commits are clear and descriptive
- [ ] Documentation updated (if needed)
```

## Development Setup

### Prerequisites

- Go 1.21 or higher
- macOS 12.0 or higher
- Git

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/macos-supply-chain-monitor.git
cd macos-supply-chain-monitor

# Add upstream remote
git remote add upstream https://github.com/straticus1/macos-supply-chain-monitor.git

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Run linters
make lint
```

### Project Structure

```
supply-chain-monitor/
├── cmd/              # CLI commands
├── internal/         # Internal packages
│   ├── daemon/      # Monitoring daemon
│   ├── db/          # Database layer
│   ├── watcher/     # File system watchers
│   └── scanner/     # Package scanners
├── main.go          # Entry point
└── go.mod           # Dependencies
```

### Coding Guidelines

#### Go Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Run `go vet` to catch common mistakes
- Write idiomatic Go code

#### Naming Conventions

- Use MixedCaps for exported names
- Use camelCase for unexported names
- Keep names concise but descriptive
- Avoid abbreviations unless they're common (e.g., `ID`, `URL`)

#### Error Handling

- Always handle errors explicitly
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Log errors at appropriate levels
- Return errors to caller when appropriate

#### Testing

- Write table-driven tests when possible
- Use meaningful test names: `TestFunctionName_Scenario`
- Aim for >80% code coverage
- Test edge cases and error conditions

Example:
```go
func TestCalculateRiskScore_UnsignedBinary(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        manager  string
        expected int
    }{
        {
            name:     "unsigned executable",
            path:     "/usr/local/bin/suspicious",
            manager:  "homebrew",
            expected: 40,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            score := calculateRiskScore(tt.path, tt.manager)
            if score != tt.expected {
                t.Errorf("got %d, want %d", score, tt.expected)
            }
        })
    }
}
```

## Adding Support for New Package Managers

To add support for a new package manager:

1. **Create a scanner** in `internal/scanner/scanner.go`:
   ```go
   func (s *Scanner) scanYourManager() (int, error) {
       // Implementation
   }
   ```

2. **Add watcher logic** in `internal/daemon/daemon.go`:
   ```go
   func (d *Daemon) detectPackageManager(path string) string {
       // Add detection logic
   }
   ```

3. **Update configuration** in `config.example.yaml`

4. **Add tests** for your new scanner

5. **Update documentation** in README.md

## Documentation

- Keep README.md up to date
- Document all exported functions and types
- Use godoc-style comments
- Update QUICKSTART.md for user-facing changes

## Community

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Discord**: Real-time chat (link in README)

## Recognition

Contributors will be recognized in:
- README.md contributors section
- Release notes
- Our gratitude 🙏

## Questions?

Don't hesitate to ask! Open an issue with the `question` label or reach out on Discord.

---

**Thank you for contributing to making supply-chain security better for everyone!** 🔒
