# Brokli

A tool for parsing and analyzing HTML links.

## Development

### Running Tests

Run all unit tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -v -race -coverprofile=coverage.out ./...
```

### Code Quality

#### Formatting
Format all Go code:
```bash
go fmt ./...
```

Check formatting without making changes:
```bash
gofmt -s -l .
```

#### Linting
Lint code using golangci-lint:
```bash
golangci-lint run ./...
```

### CI/CD

This project uses GitHub Actions for continuous integration. On every pull request, the following checks are performed:

1. **Unit Tests**: All unit tests are run with race detection and coverage reporting
2. **Format Check**: Ensures all Go code is properly formatted with `gofmt`
3. **Lint Check**: Runs `golangci-lint` to catch common issues and enforce best practices

The workflow configuration can be found in `.github/workflows/pr-checks.yml`.

### Building

Build the project:
```bash
go build .
```

This will create a `brokli` binary in the current directory.

### Usage

Run the CLI tool:
```bash
./brokli
```

Check a URL:
```bash
./brokli check url https://example.com
```

Check a sitemap:
```bash
./brokli check sitemap https://example.com/sitemap.xml
```