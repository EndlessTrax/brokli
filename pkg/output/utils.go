package output

import (
	"fmt"

	"github.com/fatih/color"

	"github.com/endlesstrax/brokli/pkg/link"
)

// getStatusIcon returns an icon based on the HTTP status code
func getStatusIcon(status int) string {
	if status == -1 {
		return "⚠️" // Warning for unchecked
	} else if status >= 200 && status < 300 {
		return "✓" // Success
	} else if status >= 300 && status < 400 {
		return "→" // Redirect
	} else if status >= 400 {
		return "✗" // Error
	}
	return "?" // Unknown
}

// getColoredStatus returns a colored status code string
func getColoredStatus(status int) string {
	statusStr := fmt.Sprintf("[%d]", status)
	if status == -1 {
		return color.YellowString(statusStr)
	} else if status >= 200 && status < 300 {
		return color.GreenString(statusStr)
	} else if status >= 300 && status < 400 {
		return color.CyanString(statusStr)
	} else if status >= 400 && status < 500 {
		return color.RedString(statusStr)
	} else if status >= 500 {
		return color.New(color.FgRed, color.Bold).Sprint(statusStr)
	}
	return statusStr
}

// isBrokenLink returns true if the status code indicates a broken link
func isBrokenLink(status int) bool {
	// A link is broken if it's unchecked (-1) or has a 4xx/5xx status code
	return status == -1 || status >= 400
}

// countBrokenLinks counts the number of broken links in a slice of AnchorTags
func countBrokenLinks(links []*link.AnchorTag) int {
	count := 0
	for _, linkPtr := range links {
		if isBrokenLink(linkPtr.Status) {
			count++
		}
	}
	return count
}

// countBrokenSitemapUrls counts the number of broken URLs in a slice of SitemapUrls
func countBrokenSitemapUrls(urls []*link.SitemapUrl) int {
	count := 0
	for _, urlPtr := range urls {
		if isBrokenLink(urlPtr.Status) {
			count++
		}
	}
	return count
}
