package output

import (
	"fmt"
	"io"
	"os"

	"github.com/endlesstrax/brokli/pkg/link"
)

// GitHubFormatter outputs GitHub Actions compatible format with workflow annotations
type GitHubFormatter struct{}

// FormatPageResults formats page results as GitHub Actions workflow commands
func (f *GitHubFormatter) FormatPageResults(writer io.Writer, results *link.PageResults, links []*link.AnchorTag) error {
	brokenCount := countBrokenLinks(links)

	// Print workflow annotations for broken links
	for _, linkPtr := range links {
		if isBrokenLink(linkPtr.Status) {
			printGitHubAnnotation(writer, linkPtr.AbsoluteUrl.String(), linkPtr.Text, linkPtr.Status)
		}
	}

	// Write summary to GITHUB_OUTPUT if available
	if err := writeGitHubOutput(brokenCount, len(links)); err != nil {
		fmt.Fprintf(writer, "Warning: Failed to write to GITHUB_OUTPUT: %v\n", err)
	}

	// Print summary to stdout
	fmt.Fprintf(writer, "Found %d broken links out of %d total\n", brokenCount, len(links))

	return nil
}

// FormatSitemapResults formats sitemap results as GitHub Actions workflow commands
func (f *GitHubFormatter) FormatSitemapResults(writer io.Writer, results *link.SitemapResults, urls []*link.SitemapUrl) error {
	brokenCount := countBrokenSitemapUrls(urls)

	// Print workflow annotations for broken URLs
	for _, urlPtr := range urls {
		if isBrokenLink(urlPtr.Status) {
			printGitHubAnnotation(writer, urlPtr.AbsoluteUrl.String(), "", urlPtr.Status)
		}
	}

	// Write summary to GITHUB_OUTPUT if available
	if err := writeGitHubOutput(brokenCount, len(urls)); err != nil {
		fmt.Fprintf(writer, "Warning: Failed to write to GITHUB_OUTPUT: %v\n", err)
	}

	// Print summary to stdout
	fmt.Fprintf(writer, "Found %d broken URLs out of %d total\n", brokenCount, len(urls))

	return nil
}

// printGitHubAnnotation outputs a GitHub Actions workflow annotation for a broken link
func printGitHubAnnotation(writer io.Writer, url, text string, status int) {
	// Use ::error for broken links to make them highly visible in GitHub Actions
	annotationType := "error"
	title := fmt.Sprintf("Broken Link (Status %d)", status)
	message := fmt.Sprintf("Link to %s returned status %d", url, status)
	if text != "" {
		message = fmt.Sprintf("Link '%s' to %s returned status %d", text, url, status)
	}
	fmt.Fprintf(writer, "::%s title=\"%s\"::%s\n", annotationType, title, message)
}

// writeGitHubOutput writes summary data to GITHUB_OUTPUT if the environment variable is set
func writeGitHubOutput(brokenCount, totalCount int) error {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile == "" {
		// GITHUB_OUTPUT not set, skip writing
		return nil
	}

	// #nosec G304,G703 -- GITHUB_OUTPUT is provided by the GitHub Actions runtime.
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_OUTPUT file: %w", err)
	}
	defer f.Close()

	// Write summary statistics as GitHub Actions step outputs
	_, err = fmt.Fprintf(f, "broken_links_count=%d\n", brokenCount)
	if err != nil {
		return fmt.Errorf("failed to write broken_links_count: %w", err)
	}

	_, err = fmt.Fprintf(f, "total_links_count=%d\n", totalCount)
	if err != nil {
		return fmt.Errorf("failed to write total_links_count: %w", err)
	}

	hasBrokenLinks := "false"
	if brokenCount > 0 {
		hasBrokenLinks = "true"
	}
	_, err = fmt.Fprintf(f, "has_broken_links=%s\n", hasBrokenLinks)
	if err != nil {
		return fmt.Errorf("failed to write has_broken_links: %w", err)
	}

	return nil
}
