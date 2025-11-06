# Contributing to Brokli

Thank you for your interest in contributing to Brokli! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Commit Convention](#commit-convention)
- [Project Structure](#project-structure)

## Code of Conduct

This project follows a simple code of conduct:

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Assume positive intent

## Getting Started

### Prerequisites

- Go 1.24.0 or higher
- [Task](https://taskfile.dev/) (optional but recommended)
- Git
- A code editor (VS Code recommended)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/brokli.git
cd brokli
```

3. Add the upstream repository:

```bash
git remote add upstream https://github.com/endlesstrax/brokli.git
```

4. Verify your remotes:

```bash
git remote -v
```

### Install Dependencies

```bash
go mod download
```

### Run Tests

Make sure everything works:

```bash
task test
# or
go test ./...
```

## Development Workflow

### 1. Create a Branch

Always work on a feature branch:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

Branch naming conventions:

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test additions or improvements

### 2. Make Changes

Write your code following the [Coding Standards](#coding-standards) section below.

### 3. Run Checks Locally

Before committing, run all CI checks locally:

```bash
task ci
```

This runs:

1. Tests with race detection
2. Format check
3. Linting

Fix any issues before proceeding.

### 4. Commit Changes

Follow the [Commit Convention](#commit-convention) below:

```bash
git add .
git commit -m "feat: add caching support for HTTP checks"
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Coding Standards

### Go Code Style

1. **Formatting**: Always run `task fmt` before committing

```bash
task fmt
# or
go fmt ./...
```

2. **Linting**: Fix all linter warnings

```bash
task ci-lint
# or
golangci-lint run ./...
```

3. **Error Handling**: Always wrap errors with context

```go
// Good
if err != nil {
    return fmt.Errorf("failed to parse URL '%s': %w", urlStr, err)
}

// Bad
if err != nil {
    return err
}
```

4. **Package Imports**: Group imports logically

```go
import (
    // Standard library
    "fmt"
    "net/http"
    
    // External packages
    "github.com/fatih/color"
    "github.com/spf13/cobra"
    
    // Internal packages
    "github.com/endlesstrax/brokli/pkg/checker"
    "github.com/endlesstrax/brokli/pkg/link"
)
```

5. **Naming Conventions**:
   - Use descriptive names: `GetHTML()` not `Get()`
   - Exported functions: `PascalCase`
   - Unexported functions: `camelCase`
   - Constants: `PascalCase` or `UPPER_SNAKE_CASE` for package-level

### Architecture Patterns

Follow the established architecture:

- `cmd/`: CLI commands and user interaction only
- `pkg/link/`: Pure data structures, no business logic
- `pkg/fetcher/`: HTTP operations only
- `pkg/parser/`: HTML/XML parsing only
- `pkg/resolver/`: URL resolution only
- `pkg/checker/`: HTTP status checking with concurrency

**Separation of Concerns**: Keep packages focused on a single responsibility.

### Error Patterns

- Constructors validate inputs: `return AnchorTag{}, fmt.Errorf("tag cannot be nil")`
- CLI commands print errors and call `cmd.Help()` for missing arguments
- Use `// #nosec G107 -- URL is provided by user` to ignore false positive gosec warnings

### Concurrency Patterns

- Use worker pools for concurrent operations
- Invoke callbacks serially (not in goroutines) to avoid thread-safety issues
- Use `sync/atomic` for counters in concurrent code
- Always test concurrent code with race detection: `go test -race`

## Testing

### Writing Tests

1. **Test file naming**: Mirror source files (`fetcher_test.go`, `checker_test.go`)

2. **Table-driven tests** for multiple scenarios:

```go
func TestResolveAbsoluteUrl(t *testing.T) {
    tests := []struct {
        name     string
        base     string
        href     string
        expected string
        wantErr  bool
    }{
        {
            name:     "absolute URL",
            base:     "https://example.com",
            href:     "https://other.com",
            expected: "https://other.com",
            wantErr:  false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ResolveAbsoluteUrl(tt.base, tt.href)
            // Assertions...
        })
    }
}
```

3. **Test nil inputs explicitly** - Most functions should error with "cannot be nil"

4. **Use httptest.NewServer()** for mocking HTTP responses:

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
}))
defer server.Close()
```

5. **Integration tests** at end of test files to combine multiple packages

### Running Tests

```bash
# Quick test
task test

# With coverage
task test-coverage

# Specific package
go test ./pkg/checker/...

# With race detection
go test -race ./...

# Verbose output
go test -v ./...
```

### Test Coverage

- Aim for 90%+ coverage on new code
- Critical packages (checker, parser, resolver) should have 95%+ coverage
- View coverage report:

```bash
task test-coverage
# Opens coverage.html in browser
```

## Submitting Changes

### Pull Request Process

1. **Update Documentation**: If adding a user-facing feature, update `README.md`

2. **Add Tests**: All new code must have tests

3. **Run CI Checks**: Ensure `task ci` passes locally

4. **Write Clear PR Description**:
   - What does this PR do?
   - Why is this change needed?
   - How was it tested?
   - Screenshots (if UI changes)

5. **Link Related Issues**: Use keywords like "Fixes #123" or "Relates to #456"

### PR Template

```markdown
## Description
Brief description of what this PR does.

## Changes
- Added feature X
- Fixed bug Y
- Refactored Z

## Testing
- Added tests in `pkg/checker/checker_test.go`
- All tests pass with race detection
- Manually tested with `./brokli check url https://example.com`

## Related Issues
Fixes #123
```

### Review Process

- PRs require at least one approval
- Address all review comments
- Keep PRs focused - one feature/fix per PR
- Be responsive to feedback

## Commit Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation only changes
- `style:` - Code style changes (formatting, missing semi-colons, etc.)
- `refactor:` - Code changes that neither fix bugs nor add features
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks (dependencies, CI, etc.)
- `perf:` - Performance improvements

### Examples

```bash
feat(checker): add retry logic for failed requests
fix(parser): handle empty href attributes correctly
docs: update README with verbose flag examples
test(resolver): add tests for special link types
chore: upgrade golangci-lint to v2.0.0
```

### Scope

Optional but recommended. Scope indicates which package or component is affected:

- `checker`
- `fetcher`
- `parser`
- `resolver`
- `link`
- `cmd`
- `ci`

## Project Structure

Understanding the codebase:

```text
brokli/
├── cmd/                    # CLI commands (Cobra)
│   ├── check.go           # check url/sitemap commands
│   └── root.go            # Root command setup
├── pkg/                   # Core packages
│   ├── checker/           # HTTP status checking
│   │   ├── checker.go    # Worker pool implementation
│   │   └── checker_test.go
│   ├── fetcher/           # HTTP operations
│   │   ├── fetcher.go    # GetHTML function
│   │   └── fetcher_test.go
│   ├── link/              # Data structures
│   │   ├── link.go       # AnchorTag, SitemapUrl, Results
│   │   └── link_test.go
│   ├── parser/            # HTML/XML parsing
│   │   ├── html.go       # HTML parsing
│   │   ├── sitemap.go    # Sitemap parsing
│   │   └── *_test.go
│   └── resolver/          # URL resolution
│       ├── resolver.go   # ResolveAbsoluteUrl, IsSpecialLink
│       └── resolver_test.go
├── .github/
│   ├── copilot-instructions.md  # AI agent guidance
│   └── workflows/        # GitHub Actions
├── main.go               # Entry point
├── Taskfile.yml          # Task definitions
├── README.md             # User documentation
├── ROADMAP.md            # Planned features
└── CONTRIBUTING.md       # This file
```

### Package Responsibilities

- **cmd/**: User interaction, flag parsing, output formatting
- **pkg/link/**: Pure data types, no business logic
- **pkg/fetcher/**: HTTP fetching only
- **pkg/parser/**: Parsing HTML/XML into data structures
- **pkg/resolver/**: URL resolution and validation
- **pkg/checker/**: Concurrent HTTP checking with worker pools

### Adding New Features

When adding a feature from the [ROADMAP.md](ROADMAP.md):

1. Check if it requires new packages or extends existing ones
2. Follow the separation of concerns pattern
3. Update `.github/copilot-instructions.md` if adding patterns
4. Update `README.md` with user-facing changes
5. Mark feature as complete in `ROADMAP.md`

## Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: Open a GitHub Issue
- **Features**: Check ROADMAP.md first, then open an Issue for discussion

## Recognition

Contributors will be recognized in:

- GitHub contributors page
- Release notes for significant contributions
- README acknowledgments section (coming soon)

Thank you for contributing to Brokli! 🥦
