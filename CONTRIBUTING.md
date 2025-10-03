# Contributing to FullMCP

Thank you for your interest in contributing to FullMCP! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please:

- Be respectful and considerate
- Use welcoming and inclusive language
- Accept constructive criticism gracefully
- Focus on what's best for the community
- Show empathy towards other contributors

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- golangci-lint (for linting)

### Fork and Clone

```bash
# Fork the repository on GitHub

# Clone your fork
git clone https://github.com/YOUR_USERNAME/fullmcp.git
cd fullmcp

# Add upstream remote
git remote add upstream https://github.com/jmcarbo/fullmcp.git
```

### Setup Development Environment

```bash
# Install dependencies
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests to verify setup
go test ./...
```

## Development Workflow

### 1. Create a Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### Branch Naming Conventions

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test improvements
- `perf/` - Performance improvements

### 2. Make Changes

Follow our [coding standards](#coding-standards) and write tests for new functionality.

### 3. Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run integration tests
go test -v -run=TestIntegration ./...
```

### 4. Run Linters

```bash
# Run golangci-lint
golangci-lint run

# Format code
gofmt -w .
goimports -w .
```

### 5. Commit Changes

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat: add new feature description"
```

#### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `perf` - Performance improvements
- `chore` - Maintenance tasks

**Examples:**
```
feat: add WebSocket transport support
fix: handle nil pointer in resource manager
docs: update authentication guide
test: add integration tests for proxy server
perf: optimize JSON schema generation
```

### 6. Push Changes

```bash
# Push to your fork
git push origin feature/your-feature-name
```

### 7. Create Pull Request

- Go to GitHub and create a Pull Request
- Fill out the PR template
- Link any related issues
- Request review from maintainers

## Coding Standards

### General Guidelines

1. **Follow Go conventions**
   - Use `gofmt` for formatting
   - Follow [Effective Go](https://golang.org/doc/effective_go.html)
   - Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

2. **Write idiomatic Go**
   ```go
   // ✅ Good
   if err != nil {
       return err
   }

   // ❌ Bad
   if err == nil {
       // continue
   } else {
       return err
   }
   ```

3. **Keep functions small**
   - Aim for functions under 50 lines
   - Single responsibility principle
   - Extract complex logic into helper functions

4. **Use meaningful names**
   ```go
   // ✅ Good
   func (tm *ToolManager) CallTool(ctx context.Context, name string, args json.RawMessage) (interface{}, error)

   // ❌ Bad
   func (tm *ToolManager) CT(c context.Context, n string, a json.RawMessage) (interface{}, error)
   ```

### Package Organization

```
fullmcp/
├── mcp/           # Core protocol types
├── server/        # Server implementation
├── client/        # Client implementation
├── builder/       # Builder APIs
├── transport/     # Transport implementations
├── auth/          # Authentication providers
├── internal/      # Internal packages
└── examples/      # Example applications
```

### Error Handling

```go
// ✅ Good: Wrap errors with context
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ✅ Good: Use MCP errors for protocol errors
if tool == nil {
    return &mcp.Error{
        Code:    mcp.ErrorCodeMethodNotFound,
        Message: fmt.Sprintf("tool %s not found", name),
    }
}

// ❌ Bad: Ignore errors
doSomething() // error ignored
```

### Concurrency

```go
// ✅ Good: Use mutexes for shared state
type Manager struct {
    items map[string]*Item
    mu    sync.RWMutex
}

func (m *Manager) Get(key string) *Item {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.items[key]
}

// ✅ Good: Pass context for cancellation
func (m *Manager) Process(ctx context.Context, data []byte) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process
    }
}
```

### Documentation

```go
// ✅ Good: Document exported types and functions
// ToolManager manages the registration and execution of tools.
// It provides thread-safe access to tools and handles tool invocation
// with automatic input validation.
type ToolManager struct {
    tools map[string]*Tool
    mu    sync.RWMutex
}

// AddTool registers a new tool with the manager.
// The tool name must be unique. If a tool with the same name
// already exists, it will be replaced.
func (tm *ToolManager) AddTool(tool *Tool) {
    // ...
}
```

## Testing Guidelines

### Test Coverage

- Aim for 95%+ test coverage
- All new features must include tests
- Bug fixes must include regression tests

### Test Structure

```go
func TestToolManager_CallTool(t *testing.T) {
    // Arrange
    tm := NewToolManager()
    tool := createTestTool()
    tm.AddTool(tool)

    // Act
    result, err := tm.CallTool(context.Background(), "test", json.RawMessage(`{}`))

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "valid",
            want:    true,
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "",
            want:    false,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Validate(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Integration Tests

```go
func TestIntegration_ClientServer(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup server
    srv := setupTestServer()
    defer srv.Close()

    // Setup client
    client := setupTestClient()
    defer client.Close()

    // Test end-to-end flow
    // ...
}
```

### Benchmarks

```go
func BenchmarkToolCall(b *testing.B) {
    tm := NewToolManager()
    tm.AddTool(createTestTool())
    args := json.RawMessage(`{"a": 5, "b": 3}`)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        tm.CallTool(context.Background(), "test", args)
    }
}
```

## Documentation

### Code Documentation

- Document all exported types, functions, and methods
- Use complete sentences
- Include examples for complex functionality
- Keep documentation up-to-date with code changes

### User Documentation

Update relevant documentation in `docs/`:

- `architecture.md` - Architecture changes
- `tools.md` - Tool-related features
- `resources.md` - Resource-related features
- `prompts.md` - Prompt-related features
- `authentication.md` - Auth changes
- `transports.md` - Transport changes
- `middleware.md` - Middleware changes

### Examples

Add examples to `examples/` directory when adding new features:

```
examples/
├── basic-server/
├── advanced-server/
├── http-server/
└── your-new-example/
    ├── main.go
    └── README.md
```

## Pull Request Process

### Before Submitting

- [ ] All tests pass
- [ ] Code is formatted (`gofmt`, `goimports`)
- [ ] Linters pass (`golangci-lint run`)
- [ ] Documentation is updated
- [ ] Examples are added/updated if needed
- [ ] Commit messages follow conventions
- [ ] Branch is up-to-date with main

### PR Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Related Issues
Fixes #123

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass
- [ ] No new warnings
```

### Review Process

1. Maintainers will review your PR
2. Address any feedback or requested changes
3. Once approved, a maintainer will merge your PR

### After Merge

```bash
# Update your local main branch
git checkout main
git pull upstream main

# Delete feature branch
git branch -d feature/your-feature-name
git push origin --delete feature/your-feature-name
```

## Release Process

Releases are managed by maintainers:

1. Version bump following [Semantic Versioning](https://semver.org/)
2. Update CHANGELOG.md
3. Create and push tag
4. GitHub Actions builds and publishes release

### Versioning

- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes (backward compatible)

## Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: Open a GitHub Issue
- **Security**: Email security@example.com (use private disclosure)

## Recognition

Contributors will be:
- Listed in the README
- Mentioned in release notes
- Credited in the CHANGELOG

Thank you for contributing to FullMCP! 🎉
