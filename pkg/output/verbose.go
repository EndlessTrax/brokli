package output

import (
	"fmt"
	"io"

	"github.com/fatih/color"

	"github.com/endlesstrax/brokli/pkg/link"
)

// VerboseFormatter outputs all links with their status codes
type VerboseFormatter struct{}

// FormatPageResults formats page results showing all links
func (f *VerboseFormatter) FormatPageResults(writer io.Writer, results *link.PageResults, links []*link.AnchorTag) error {
	brokenCount := countBrokenLinks(links)

	// Show all links in verbose mode
	fmt.Fprintln(writer, "\nAll Links:")
	for i, linkPtr := range links {
		statusIcon := getStatusIcon(linkPtr.Status)
		coloredStatus := getColoredStatus(linkPtr.Status)
		fmt.Fprintf(writer, "%s %d. %s %s -> %s\n", statusIcon, i+1, coloredStatus, linkPtr.Text, linkPtr.AbsoluteUrl.String())
	}

	// Display summary
	fmt.Fprintln(writer)
	if brokenCount > 0 {
		color.New(color.FgRed).Fprintf(writer, "Summary: %d broken links found out of %d total\n", brokenCount, len(links))
	} else {
		color.New(color.FgGreen).Fprintf(writer, "Summary: All %d links are working\n", len(links))
	}
	fmt.Fprintln(writer)

	return nil
}

// FormatSitemapResults formats sitemap results showing all URLs
func (f *VerboseFormatter) FormatSitemapResults(writer io.Writer, results *link.SitemapResults, urls []*link.SitemapUrl) error {
	brokenCount := countBrokenSitemapUrls(urls)

	// Show all URLs in verbose mode
	fmt.Fprintln(writer, "\nAll URLs:")
	for i, urlPtr := range urls {
		statusIcon := getStatusIcon(urlPtr.Status)
		coloredStatus := getColoredStatus(urlPtr.Status)
		fmt.Fprintf(writer, "%s %d. %s %s", statusIcon, i+1, coloredStatus, urlPtr.AbsoluteUrl.String())
		if urlPtr.LastMod != "" {
			fmt.Fprintf(writer, " (modified: %s)", color.CyanString(urlPtr.LastMod))
		}
		if urlPtr.Priority != "" {
			fmt.Fprintf(writer, " (priority: %s)", color.YellowString(urlPtr.Priority))
		}
		fmt.Fprintln(writer)
	}

	// Display summary
	fmt.Fprintln(writer)
	if brokenCount > 0 {
		color.New(color.FgRed).Fprintf(writer, "Summary: %d broken URLs found out of %d total\n", brokenCount, len(urls))
	} else {
		color.New(color.FgGreen).Fprintf(writer, "Summary: All %d URLs are working\n", len(urls))
	}
	fmt.Fprintln(writer)

	return nil
}
