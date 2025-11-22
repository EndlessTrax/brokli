package output

import (
	"fmt"
	"io"

	"github.com/fatih/color"

	"github.com/endlesstrax/brokli/pkg/link"
)

// DefaultFormatter outputs only broken links with color-coded status
type DefaultFormatter struct{}

// FormatPageResults formats page results showing only broken links
func (f *DefaultFormatter) FormatPageResults(writer io.Writer, results *link.PageResults, links []*link.AnchorTag) error {
	brokenCount := countBrokenLinks(links)

	if brokenCount > 0 {
		fmt.Fprintln(writer, "\nBroken Links:")
		count := 1
		for _, linkPtr := range links {
			if isBrokenLink(linkPtr.Status) {
				statusIcon := getStatusIcon(linkPtr.Status)
				coloredStatus := getColoredStatus(linkPtr.Status)
				fmt.Fprintf(writer, "%s %d. %s %s -> %s\n", statusIcon, count, coloredStatus, linkPtr.Text, linkPtr.AbsoluteUrl.String())
				count++
			}
		}
		// Display summary for broken links
		fmt.Fprintln(writer)
		color.New(color.FgRed).Fprintf(writer, "Summary: %d broken links found out of %d total\n", brokenCount, len(links))
		fmt.Fprintln(writer)
	} else {
		color.New(color.FgGreen).Fprintln(writer, "\n✓ All links are working!")
		fmt.Fprintln(writer)
	}

	return nil
}

// FormatSitemapResults formats sitemap results showing only broken URLs
func (f *DefaultFormatter) FormatSitemapResults(writer io.Writer, results *link.SitemapResults, urls []*link.SitemapUrl) error {
	brokenCount := countBrokenSitemapUrls(urls)

	if brokenCount > 0 {
		fmt.Fprintln(writer, "\nBroken URLs:")
		count := 1
		for _, urlPtr := range urls {
			if isBrokenLink(urlPtr.Status) {
				statusIcon := getStatusIcon(urlPtr.Status)
				coloredStatus := getColoredStatus(urlPtr.Status)
				fmt.Fprintf(writer, "%s %d. %s %s", statusIcon, count, coloredStatus, urlPtr.AbsoluteUrl.String())
				if urlPtr.LastMod != "" {
					fmt.Fprintf(writer, " (modified: %s)", color.CyanString(urlPtr.LastMod))
				}
				if urlPtr.Priority != "" {
					fmt.Fprintf(writer, " (priority: %s)", color.YellowString(urlPtr.Priority))
				}
				fmt.Fprintln(writer)
				count++
			}
		}
		// Display summary for broken URLs
		fmt.Fprintln(writer)
		color.New(color.FgRed).Fprintf(writer, "Summary: %d broken URLs found out of %d total\n", brokenCount, len(urls))
		fmt.Fprintln(writer)
	} else {
		color.New(color.FgGreen).Fprintln(writer, "\n✓ All URLs are working!")
		fmt.Fprintln(writer)
	}

	return nil
}
