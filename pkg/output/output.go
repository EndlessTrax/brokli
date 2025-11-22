package output

import (
	"io"

	"github.com/endlesstrax/brokli/pkg/link"
)

// Format represents an output format type
type Format string

const (
	// FormatDefault is the standard terminal output showing only broken links
	FormatDefault Format = "default"
	// FormatVerbose is terminal output showing all links
	FormatVerbose Format = "verbose"
	// FormatGitHub is GitHub Actions compatible output with annotations
	FormatGitHub Format = "github"
)

// String constants for format names (for use in CLI flags and comparisons)
const (
	FormatNameDefault = "default"
	FormatNameVerbose = "verbose"
	FormatNameGitHub  = "github"
)

// Formatter is the interface for all output formatters
type Formatter interface {
	// FormatPageResults formats and writes the results of checking links on a page
	FormatPageResults(writer io.Writer, results *link.PageResults, links []*link.AnchorTag) error

	// FormatSitemapResults formats and writes the results of checking a sitemap
	FormatSitemapResults(writer io.Writer, results *link.SitemapResults, urls []*link.SitemapUrl) error
}

// NewFormatter creates a formatter based on the specified format type
func NewFormatter(format Format) Formatter {
	switch format {
	case FormatGitHub:
		return &GitHubFormatter{}
	case FormatVerbose:
		return &VerboseFormatter{}
	default:
		return &DefaultFormatter{}
	}
}
