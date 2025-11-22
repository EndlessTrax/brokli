# Brokli Development Guide

## Project Overview

Brokli (a play on "broken links") is a CLI tool for checking broken links on websites during development. It helps developers validate all links on a page or in a sitemap by fetching URLs, checking HTTP status codes concurrently, and displaying results in a pretty terminal output.

**Architecture**: Go CLI using Cobra framework with clear separation of concerns:
- `cmd/`: CLI commands and user interaction (Cobra commands)
- `pkg/link/`: Pure data structures (`AnchorTag`, `SitemapUrl`, `PageResults`, `SitemapResults`)
- `pkg/fetcher/`: HTTP operations (`GetHTML`)
- `pkg/parser/`: HTML/XML parsing (`html.go`, `sitemap.go`)
- `pkg/resolver/`: URL resolution logic (`ResolveAbsoluteUrl`, `IsSpecialLink`)
- `pkg/output/`: Output formatting with interface-based design (`Formatter` interface, `DefaultFormatter`, `VerboseFormatter`, `GitHubFormatter`)

**Target Use Case**: Local development workflow - developers run `brokli check url https://localhost:3000` or `brokli check sitemap https://localhost:3000/sitemap.xml` to validate links before deployment.

## Completed Features

- ✅ **Concurrent HTTP Checking** - Worker pool with configurable workers (default: 10)
- ✅ **Progress Indication** - Real-time counter with thread-safe serial callback
- ✅ **Colored Terminal Output** - Status code coloring (green/red/cyan/yellow)
- ✅ **Multiple Output Formats** - `--output-format` flag with `default`, `verbose`, and `github` options
- ✅ **GitHub Actions Integration** - Native support with workflow annotations and step outputs
- ✅ **Smart Filtering** - Display broken links (4xx/5xx) by default
- ✅ **URL & Sitemap Support** - Check single pages or entire sitemaps
- ✅ **Comprehensive Testing** - 94%+ test coverage with race detection

## Key Patterns & Conventions

### Project Structure
- Entry point is `main.go` (minimal, just calls `cmd.Execute()`)
- Commands use Cobra pattern: `rootCmd` in `cmd/root.go`, subcommands in `cmd/check.go`
- Business logic separated by concern:
  - `pkg/link`: Pure data types, no business logic
  - `pkg/fetcher`: Only HTTP fetching
  - `pkg/parser`: Only parsing (HTML/XML → data structures)
  - `pkg/resolver`: Only URL resolution
  - `pkg/output`: Only output formatting (interface-based design)
- Import packages with descriptive names: `fetcher.GetHTML()`, `parser.ParseHTML()`, `resolver.ResolveAbsoluteUrl()`, `output.NewFormatter()`

### Error Handling
- Functions return errors wrapped with context: `fmt.Errorf("failed to parse URL '%s': %w", urlStr, err)`
- Constructors validate inputs and return descriptive errors: `return AnchorTag{}, fmt.Errorf("tag cannot be nil")`
- CLI commands print errors to stdout and call `cmd.Help()` for missing arguments

### Data Model
- `AnchorTag` and `SitemapUrl` both initialize `Status` field to `-1` (unknown) before HTTP checks
- Special link types (mailto:, javascript:, #fragments) resolve to empty `url.URL{}` via `resolver.IsSpecialLink()` and are skipped during status checks
- Status codes: `-1` = unchecked, `200` = OK, `404` = not found, etc.
- Data structures are pure - no HTTP logic in types (separation of concerns)

### Concurrency & Performance
- HTTP status checks use goroutines with worker pools (default: 10 concurrent workers)
- Checker configuration (`pkg/checker.Config`):
  - `MaxWorkers`: Number of concurrent HTTP requests (default: 10)
  - `Timeout`: Maximum time per request (default: 10s)
  - `UserAgent`: Custom User-Agent header (default: "Brokli/0.1.0 (Broken Link Checker)")
  - `MaxRedirects`: Maximum redirects to follow (default: 10)
  - `ProgressCallback`: Optional callback for progress updates (func(checked, total int))
- Progress tracking: Thread-safe with serial callback invocation in results loop
- Terminal output shows progress indication with colored status codes
- Rate limiting can be controlled via `MaxWorkers` to avoid overwhelming local dev servers
- Checker package operates on `link.AnchorTag` and `link.SitemapUrl` to set Status fields

### Output & Display
- **Output Formatting** uses interface-based design in `pkg/output/`:
  - `Formatter` interface defines `FormatPageResults()` and `FormatSitemapResults()`
  - Three implementations: `DefaultFormatter`, `VerboseFormatter`, `GitHubFormatter`
  - Factory function: `output.NewFormatter(format)` creates appropriate formatter
  - String constants for format names: `output.FormatNameDefault`, `output.FormatNameVerbose`, `output.FormatNameGitHub`
- **CLI Flags**:
  - `--output-format` (or `-o`): Choose output format: `default`, `verbose`, or `github`
  - `-v/--verbose`: Deprecated but maintained for backward compatibility, sets format to verbose
- **Default Format** (`output.FormatDefault`):
  - Shows only broken links (4xx/5xx status codes)
  - Color-coded status codes
  - Summary with count of broken links
- **Verbose Format** (`output.FormatVerbose`):
  - Shows all links with their status codes
  - Includes working links (2xx) and redirects (3xx)
  - Full summary statistics
- **GitHub Actions Format** (`output.FormatGitHub`):
  - Emits workflow annotations: `::error title="Broken Link (Status 404)"::Link 'text' to url returned status 404`
  - Writes to `GITHUB_OUTPUT` environment file (if set) with secure 0600 permissions
  - Step outputs: `broken_links_count`, `total_links_count`, `has_broken_links`
  - Simplified text summary for stdout
- Color-coded status using `github.com/fatih/color`:
  - Green: 2xx success codes
  - Cyan: 3xx redirect codes
  - Red: 4xx client errors
  - Bold Red: 5xx server errors
  - Yellow: -1 unchecked/error state
- Helper functions in `pkg/output/utils.go`:
  - `getStatusIcon()`: Returns emoji/symbol for status
  - `getColoredStatus()`: Returns colored status code string
  - `isBrokenLink()`: Determines if link should be displayed by default
  - `countBrokenLinks()`: Counts broken links in slice
  - `countBrokenSitemapUrls()`: Counts broken sitemap URLs

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

## Commit Conventions
Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Test additions or changes
- `chore:` - Maintenance tasks

**Examples**:
```
feat(checker): add retry logic for failed requests
fix(parser): handle empty href attributes correctly
docs: update README with verbose flag examples
test(resolver): add tests for special link types
chore: upgrade golangci-lint to v2.0.0
```

## Debugging
Use VS Code launch configurations (`.vscode/launch.json`):
- "Launch file": Runs `main.go` with args `["check", "https://your-url-here.com"]`
- "Launch Package": Debugs current package

## Common Tasks

**Adding progress tracking to new features:**
1. Use `checker.Config.ProgressCallback` for async operations
2. Invoke callback serially (not in goroutines) to avoid thread-safety issues
3. Call after each item completes: `if cfg.ProgressCallback != nil { cfg.ProgressCallback(completed, total) }`
4. Test with `TestCheckLinks_ProgressCallback` pattern

**Adding a new CLI command:**
1. Add command in `cmd/` file, attach to parent with `rootCmd.AddCommand(newCmd)`
2. Import packages with full names (`fetcher`, `parser`, `resolver`, `link`)
3. Follow pattern: check args → parse URL → call `fetcher.GetHTML()` → `parser.ParseHTML()` → process results

**Adding a new link type:**
1. Update `resolver.IsSpecialLink()` in `pkg/resolver/resolver.go` with special case logic
2. Add tests in `resolver_test.go` with expected empty URL behavior
3. Update integration tests in parser package if needed

**Adding a new output format:**
1. Create new formatter struct in `pkg/output/` (e.g., `JSONFormatter`)
2. Implement the `Formatter` interface with `FormatPageResults()` and `FormatSitemapResults()` methods
3. Add new format constant to `pkg/output/output.go` (e.g., `FormatJSON Format = "json"`, `FormatNameJSON = "json"`)
4. Update `NewFormatter()` factory function to handle new format
5. Add comprehensive tests in `output_test.go` for the new formatter
6. Update CLI flag help text in `cmd/check.go` to include new format option
7. Document the new format in README.md

**Modifying data structures:**
- Data types are in `pkg/link/link.go` - keep them pure (no business logic)
- Add validation in constructor functions in `pkg/parser/` (e.g., `newAnchorTag`, `newSitemapUrl`)
- Write tests for nil/invalid inputs and happy path

**Adding colored output to new commands:**
- Import `github.com/fatih/color` for terminal coloring
- Define color schemes: `color.New(color.FgGreen)`, `color.New(color.FgRed, color.Bold)`
- Use helper functions from `pkg/output/utils.go`: `getStatusIcon()`, `getColoredStatus()` for consistency
- Show summary statistics: total links, broken links, redirects
- Follow existing formatter pattern in `pkg/output/` for consistent UX

**Working with output formatters:**
- All formatters write to `io.Writer` (typically `os.Stdout`)
- Use `output.NewFormatter(format)` factory to get appropriate formatter instance
- Format selection logic in `cmd/check.go` uses string constants from `output` package
- Backward compatibility: `-v` flag still sets format to verbose (deprecated pattern)
- GitHub formatter handles `GITHUB_OUTPUT` environment variable automatically
