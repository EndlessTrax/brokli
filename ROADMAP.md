# Brokli Roadmap

This document outlines the planned features and enhancements for Brokli. Items are organized by priority and complexity.

## Version 0.2.0 - Configuration & Caching

### Configuration File Support

**Priority**: High  
**Complexity**: Medium

Add support for `.brokli.yml` configuration files to persist settings across runs.

**Features**:

- Load config from `.brokli.yml` in current directory or home directory
- Support for all command-line flags as configuration options
- Environment variable overrides
- Config file validation and error reporting

**Example `.brokli.yml`**:

```yaml
# HTTP Configuration
max-workers: 20
timeout: 10s
max-redirects: 5
user-agent: "MyBot/1.0"

# Output Configuration
verbose: false
color: true

# Link Checking
exclude-patterns:
  - "https://example.com/admin/*"
  - "mailto:*"
  - "#*"

# Status Code Handling
acceptable-codes:
  - 200
  - 201
  - 301
  - 302
```

### Link Caching

**Priority**: High  
**Complexity**: Medium

Cache HTTP check results to avoid re-checking unchanged links.

**Features**:

- Store results in local SQLite database or JSON file
- TTL-based expiration (default: 24 hours)
- Cache key based on URL + content hash
- `--no-cache` flag to bypass cache
- `--clear-cache` command to reset cache
- Cache statistics in output

**Benefits**:

- Faster subsequent runs
- Reduced load on target servers
- Better CI/CD integration

### Export Formats

**Priority**: Medium  
**Complexity**: Low

Output results to structured file formats for further processing.

**Formats**:

- **JSON** - Machine-readable, programmatic integration
- **CSV** - Spreadsheet import, data analysis
- **Markdown** - Documentation generation
- **HTML** - Standalone report with styling

**Usage**:

```bash
brokli check url https://example.com --export json --output results.json
brokli check sitemap https://example.com/sitemap.xml --export html --output report.html
```

**JSON Schema**:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "source": "https://example.com",
  "type": "page",
  "summary": {
    "total": 10,
    "broken": 2,
    "ok": 7,
    "redirects": 1
  },
  "links": [
    {
      "text": "Home",
      "url": "https://example.com",
      "status": 200,
      "duration_ms": 150
    }
  ]
}
```

## Version 0.3.0 - Advanced Checking

### Retry Logic

**Priority**: Medium  
**Complexity**: Low

Automatically retry failed requests before marking as broken.

**Features**:

- Configurable retry count (default: 3)
- Exponential backoff between retries
- Retry only on specific status codes (timeouts, 5xx errors)
- `--retry-count` and `--retry-delay` flags

### Custom Status Code Handlers

**Priority**: Medium  
**Complexity**: Medium

Define acceptable status codes per URL pattern.

**Features**:

- Pattern-based rules in config file
- Multiple patterns with priority
- Default fallback behavior

**Configuration**:

```yaml
status-rules:
  - pattern: "https://api.example.com/*"
    acceptable: [200, 201, 202]
  - pattern: "https://cdn.example.com/*"
    acceptable: [200, 304]
  - pattern: "https://legacy.example.com/*"
    acceptable: [200, 301, 302]
```

### Response Time Tracking

**Priority**: Low  
**Complexity**: Low

Track and report response times for each link.

**Features**:

- Display response time in verbose mode
- Flag slow links (> configurable threshold)
- Average/min/max response times in summary
- Export timing data for performance analysis

## Version 0.4.0 - Parallel & Notifications

### Parallel Sitemap Processing

**Priority**: Medium  
**Complexity**: Medium

Check multiple sitemaps concurrently.

**Features**:

- Auto-discover sitemaps from sitemap index files
- Check all discovered sitemaps in parallel
- Aggregate results across all sitemaps
- Support for sitemap index files

**Usage**:

```bash
# Check all sitemaps in sitemap index
brokli check sitemap https://example.com/sitemap_index.xml
```

### Notifications

**Priority**: Low  
**Complexity**: Medium

Send alerts when broken links are found.

**Integrations**:

- **Slack** - Webhook integration with formatted messages
- **Discord** - Webhook integration
- **Email** - SMTP configuration
- **Custom Webhooks** - Generic HTTP POST with JSON payload

**Configuration**:

```yaml
notifications:
  slack:
    webhook-url: "https://hooks.slack.com/services/..."
    channel: "#broken-links"
    mention-on-failure: "@dev-team"
  
  discord:
    webhook-url: "https://discord.com/api/webhooks/..."
  
  webhook:
    url: "https://example.com/webhook"
    headers:
      Authorization: "Bearer token123"
```

## Version 0.5.0 - Reporting & History

### HTML Report Generation

**Priority**: Medium  
**Complexity**: High

Generate beautiful, interactive HTML reports.

**Features**:

- Responsive design for mobile viewing
- Interactive filtering and sorting
- Charts and graphs (status code distribution, response times)
- Collapsible sections for large reports
- Export to standalone HTML file (no external dependencies)
- Print-friendly styling

### Historical Tracking

**Priority**: Low  
**Complexity**: High

Track broken links over time and visualize trends.

**Features**:

- Store check results in database with timestamps
- Compare results between runs
- Trend graphs (broken links over time)
- "Fixed" and "newly broken" link detection
- Historical report generation

**Commands**:

```bash
# Enable historical tracking
brokli check url https://example.com --track

# View history
brokli history show

# Compare two runs
brokli history diff --from 2024-01-01 --to 2024-01-15
```

## Version 1.0.0 - Spider & Advanced Features

### Spider Mode

**Priority**: High  
**Complexity**: High

Recursively crawl entire websites to discover and check all links.

**Features**:

- Follow links to discover new pages
- Respect robots.txt
- Configurable depth limit
- Domain restriction (stay within same domain)
- Rate limiting to avoid overwhelming servers
- Progress indication for large crawls

**Usage**:

```bash
# Crawl entire site up to 3 levels deep
brokli spider https://example.com --depth 3

# Stay within subdomain only
brokli spider https://blog.example.com --same-domain
```

### Diff Mode

**Priority**: Medium  
**Complexity**: Medium

Compare link status between two versions or environments.

**Features**:

- Compare two URLs (e.g., staging vs production)
- Compare current state vs cached results
- Highlight newly broken links
- Highlight fixed links
- Export diff report

**Usage**:

```bash
# Compare staging vs production
brokli diff https://staging.example.com https://example.com

# Compare current vs previous run
brokli diff https://example.com --cached
```

### Exclude Patterns

**Priority**: Medium  
**Complexity**: Low

Skip checking certain URL patterns.

**Features**:

- Glob pattern matching
- Regex support
- Command-line and config file support
- Common patterns as presets (social media, analytics, etc.)

**Usage**:

```bash
# Exclude patterns via CLI
brokli check url https://example.com --exclude "*/admin/*" --exclude "mailto:*"
```

**Configuration**:

```yaml
exclude-patterns:
  - "https://example.com/admin/*"
  - "https://analytics.google.com/*"
  - "https://*.facebook.com/*"
  - "mailto:*"
  - "tel:*"
  - "#*"
```

## Future Enhancements

### Plugin System

**Priority**: Low  
**Complexity**: High

Extend Brokli's functionality with custom plugins.

**Features**:

- Go plugin architecture
- Plugin discovery and loading
- Hook system for various stages (pre-check, post-check, output formatting)
- Example plugins: custom checkers, custom output formats, integrations

### Docker Image

**Priority**: Low  
**Complexity**: Low

Pre-built Docker images for easy deployment.

**Features**:

- Official Docker Hub image
- Multi-architecture support (amd64, arm64)
- Minimal Alpine-based image
- GitHub Container Registry support

**Usage**:

```bash
docker run -v $(pwd):/workspace ghcr.io/endlesstrax/brokli:latest \
  check url https://example.com
```

### GitHub Action

**Priority**: Medium  
**Complexity**: Medium

Ready-to-use GitHub Action for CI workflows.

**Features**:

- Simple YAML configuration
- Automatic comment on PRs with broken links
- Fail build on broken links
- Cache support for faster runs
- Configurable thresholds

**Usage**:

```yaml
name: Check Broken Links
on: [pull_request]
jobs:
  check-links:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: endlesstrax/brokli-action@v1
        with:
          url: https://preview.example.com
          fail-on-broken: true
```

### Watch Mode

**Priority**: Low  
**Complexity**: Medium

Continuously monitor a site for broken links.

**Features**:

- Watch mode with configurable interval
- Desktop notifications on new broken links
- File watcher for local HTML files
- Auto-reload on changes

**Usage**:

```bash
# Check every 5 minutes
brokli watch https://example.com --interval 5m

# Watch local files
brokli watch ./dist/index.html --local
```

## Community Requests

This section tracks feature requests from the community. If you have an idea, please open an issue on GitHub!

---

**Note**: This roadmap is subject to change based on community feedback and project priorities. Dates are estimates and may shift as development progresses.

## Contributing

Want to help implement a feature from the roadmap? Check out our [Contributing Guide](CONTRIBUTING.md) and look for issues labeled `roadmap` or `help-wanted` on GitHub!
