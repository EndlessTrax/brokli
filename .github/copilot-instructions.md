# Brokli Development Guide

## Project Overview

Brokli (a play on "broken links") is a CLI tool for checking broken links on websites during development. It helps developers validate all links on a page or in a sitemap by fetching URLs, checking HTTP status codes concurrently, and displaying results in a pretty terminal output.

**Current Status**: Core parsing and URL resolution complete. Next: implement concurrent HTTP checking and pretty terminal output.

**Architecture**: Go CLI using Cobra framework with clear separation of concerns:
- `cmd/`: CLI commands and user interaction (Cobra commands)
- `pkg/link/`: Pure data structures (`AnchorTag`, `SitemapUrl`, `PageResults`, `SitemapResults`)
- `pkg/fetcher/`: HTTP operations (`GetHTML`)
- `pkg/parser/`: HTML/XML parsing (`html.go`, `sitemap.go`)
- `pkg/resolver/`: URL resolution logic (`ResolveAbsoluteUrl`, `IsSpecialLink`)

**Target Use Case**: Local development workflow - developers run `brokli check url https://localhost:3000` or `brokli check sitemap https://localhost:3000/sitemap.xml` to validate links before deployment.

## Key Patterns & Conventions

### Project Structure
- Entry point is `main.go` (minimal, just calls `cmd.Execute()`)
- Commands use Cobra pattern: `rootCmd` in `cmd/root.go`, subcommands in `cmd/check.go`
- Business logic separated by concern:
  - `pkg/link`: Pure data types, no business logic
  - `pkg/fetcher`: Only HTTP fetching
  - `pkg/parser`: Only parsing (HTML/XML → data structures)
  - `pkg/resolver`: Only URL resolution
- Import packages with descriptive names: `fetcher.GetHTML()`, `parser.ParseHTML()`, `resolver.ResolveAbsoluteUrl()`

### Error Handling
- Functions return errors wrapped with context: `fmt.Errorf("failed to parse URL '%s': %w", urlStr, err)`
- Constructors validate inputs and return descriptive errors: `return AnchorTag{}, fmt.Errorf("tag cannot be nil")`
- CLI commands print errors to stdout and call `cmd.Help()` for missing arguments

### Data Model
- `AnchorTag` and `SitemapUrl` both initialize `Status` field to `-1` (unknown) before HTTP checks
- Special link types (mailto:, javascript:, #fragments) resolve to empty `url.URL{}` via `resolver.IsSpecialLink()` and are skipped during status checks
- Status codes: `-1` = unchecked, `200` = OK, `404` = not found, etc.
- Data structures are pure - no HTTP logic in types (separation of concerns)

### Concurrency & Performance (Future)
- HTTP status checks should use goroutines with worker pools to check links concurrently
- Consider rate limiting to avoid overwhelming local dev servers
- Terminal output should show progress (e.g., "Checking 45/120 links...")
- Checker package will operate on `link.AnchorTag` and `link.SitemapUrl` to set Status fields

### Testing Patterns
- Test files mirror source files: `fetcher_test.go`, `resolver_test.go`, `html_test.go`, `sitemap_test.go`
- Helper functions: `createTestNode()`, `createTextNode()` for building HTML nodes in parser tests
- Integration tests at end of test files combine multiple packages (e.g., `TestIntegration` in `html_test.go`)
- Use `httptest.NewServer()` for mocking HTTP responses in tests
- Test nil inputs explicitly - most functions should error with "cannot be nil" message

## Development Workflow

### Build & Run
Use Taskfile (not Makefile) for all commands:
```bash
task build          # Build binary
task run            # Build and run (shows help)
task run-check-url  # Example: check URL
task dev            # Full dev cycle: deps, test, fmt, build
```

### Testing
```bash
task test               # Quick test run
task test-coverage      # Coverage report → coverage.html
task ci                 # Run all CI checks locally
```

### Code Quality
- **Formatting**: Always run `task fmt` before committing
- **Linting**: Uses golangci-lint v2 config (`.golangci.yml`) with gosec, goconst, misspell
- **gosec**: Ignore false positives with `// #nosec G107 -- URL is provided by user` pattern
- Test files excluded from gosec/unparam checks (see `.golangci.yml` exclusions)

## CI/CD
GitHub Actions runs on PRs:
1. **Unit Tests**: Race detection + coverage upload to Codecov
2. **Format Check**: Fails if `gofmt -s -l .` returns files
3. **Lint**: golangci-lint with 5m timeout

Match CI locally with: `task ci` or individual tasks like `task ci-test`, `task ci-fmt-check`, `task ci-lint`

## Debugging
Use VS Code launch configurations (`.vscode/launch.json`):
- "Launch file": Runs `main.go` with args `["check", "https://rickywhite.net"]`
- "Launch Package": Debugs current package

## Common Tasks

**Implementing concurrent HTTP checks:**
1. Create `pkg/checker/` package with worker pool pattern
2. Use channels to distribute `link.AnchorTag` and `link.SitemapUrl` to worker goroutines
3. Workers call `http.Head()` on `link.AbsoluteUrl` and update `Status` fields
4. Add timeout context to prevent hanging on slow/dead links

**Adding a new CLI command:**
1. Add command in `cmd/` file, attach to parent with `rootCmd.AddCommand(newCmd)`
2. Import packages with full names (`fetcher`, `parser`, `resolver`, `link`)
3. Follow pattern: check args → parse URL → call `fetcher.GetHTML()` → `parser.ParseHTML()` → process results

**Adding a new link type:**
1. Update `resolver.IsSpecialLink()` in `pkg/resolver/resolver.go` with special case logic
2. Add tests in `resolver_test.go` with expected empty URL behavior
3. Update integration tests in parser package if needed

**Modifying data structures:**
- Data types are in `pkg/link/link.go` - keep them pure (no business logic)
- Add validation in constructor functions in `pkg/parser/` (e.g., `newAnchorTag`, `newSitemapUrl`)
- Write tests for nil/invalid inputs and happy path

**Pretty terminal output:**
- Consider libraries like `github.com/fatih/color` for colored output (green=200, red=404, yellow=warnings)
- Use `github.com/olekukonko/tablewriter` for tabular link status display
- Show summary statistics: total links, broken links, warnings
