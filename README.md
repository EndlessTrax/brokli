# Brokli 🥦

> A fast, concurrent broken link checker for websites and sitemaps

Brokli (a play on "broken links") is a CLI tool that helps developers validate all links on a page or sitemap during development. It checks HTTP status codes concurrently and displays results with color-coded output, making it easy to spot broken links before deployment.

## Features

- 🚀 **Concurrent Checking** - Uses worker pools to check multiple links simultaneously (default: 10 workers)
- 🎨 **Color-Coded Output** - Green for success, red for errors, cyan for redirects
- 📊 **Progress Indication** - Real-time progress counter shows checking status
- 🎯 **Smart Filtering** - Shows only broken links by default to reduce noise
- 🔍 **Verbose Mode** - Optional flag to display all links with their status codes
- ⚙️ **Configurable** - Adjust workers, timeouts, redirects, and user agent
- 🌐 **Sitemap Support** - Check entire sitemaps with metadata display
- 🧪 **Well Tested** - 94%+ test coverage with comprehensive test suite

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/EndlessTrax/brokli.git
cd brokli

# Build the binary
go build -o brokli .

# Optionally, move to your PATH
sudo mv brokli /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/endlesstrax/brokli@latest
```

## Usage

### Check a Single Page

Check all links on a webpage and show only broken ones:

```bash
brokli check url https://example.com
```

**Output:**

```text
Found 10 links
Checking links... 10/10

✓ All links are working!
```

### Check with Verbose Output

Show all links with their status codes:

```bash
brokli check url https://example.com --verbose
# or use the short flag
brokli check url https://example.com -v
```

**Output:**

```text
Found 10 links
Checking links... 10/10

All Links:
✓ 1. [200] Home -> https://example.com
✓ 2. [200] About -> https://example.com/about
→ 3. [301] Old Page -> https://example.com/redirect
✗ 4. [404] Missing -> https://example.com/missing

Summary: 1 broken links found out of 10 total
```

### Check a Sitemap

Check all URLs in a sitemap:

```bash
brokli check sitemap https://example.com/sitemap.xml
```

**Output:**

```text
Found 70 URLs in sitemap
Checking URLs... 70/70

Broken URLs:
✗ 1. [404] https://example.com/old-page (modified: 2023-09-24T00:00:00+00:00)
✗ 2. [404] https://example.com/removed (modified: 2024-01-15T00:00:00+00:00)

Summary: 2 broken URLs found out of 70 total
```

### Status Code Colors

- 🟢 **Green** `[200]` - Success (2xx status codes)
- 🔵 **Cyan** `[301]` - Redirects (3xx status codes)
- 🔴 **Red** `[404]` - Client errors (4xx status codes)
- 🔴 **Bold Red** `[500]` - Server errors (5xx status codes)
- 🟡 **Yellow** `[-1]` - Unchecked/errors

## Configuration

Brokli uses sensible defaults but can be configured via code (configuration file support coming soon):

```go
config := checker.DefaultConfig()
config.MaxWorkers = 20          // Concurrent workers (default: 10)
config.Timeout = 5 * time.Second // Request timeout (default: 10s)
config.MaxRedirects = 5          // Max redirects to follow (default: 10)
config.UserAgent = "MyBot/1.0"   // Custom user agent
```

## Use Cases

### Local Development

Validate links on your local dev server before pushing:

```bash
brokli check url https://localhost:3000
brokli check sitemap https://localhost:3000/sitemap.xml
```

### Pre-Deployment Checks

Quick validation before deploying to production:

```bash
# Check staging site
brokli check sitemap https://staging.example.com/sitemap.xml -v
```

## Roadmap

### Planned Features

- [ ] **Configuration File Support** - `.brokli.yml` for persistent settings
- [ ] **Link Caching** - Cache results to avoid re-checking unchanged links
- [ ] **Export Formats** - Output results to JSON, CSV, or Markdown
- [ ] **Multi-Sitemap Processing** - Check multiple sitemaps concurrently

## Development

### Prerequisites

- Go 1.24.0 or higher
- [Task](https://taskfile.dev/) (optional, for running tasks)

### Running Tests

Run all unit tests:

```bash
task test
# or
go test ./...
```

Run tests with coverage:

```bash
task test-coverage
# or
go test -v -race -coverprofile=coverage.out ./...
```

### Code Quality

#### Formatting

Format all Go code:

```bash
task fmt
# or
go fmt ./...
```

Check formatting without making changes:

```bash
gofmt -s -l .
```

#### Linting

Lint code using golangci-lint:

```bash
task ci-lint
# or
golangci-lint run ./...
```

### Running All CI Checks Locally

Before pushing, run all CI checks:

```bash
task ci
```

This runs:

1. Tests with race detection
2. Format check
3. Linting

### Building

Build the project:

```bash
task build
# or
go build -o brokli .
```

This creates a `brokli` binary in the current directory.

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on:

- Development workflow and branch naming conventions
- Coding standards and architecture patterns
- Testing guidelines
- Pull request process
- Commit message conventions

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Colored output using [fatih/color](https://github.com/fatih/color)
- HTML parsing with [golang.org/x/net/html](https://pkg.go.dev/golang.org/x/net/html)

---

Made with ❤️ for developers who care about link health
