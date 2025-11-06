package cmd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/endlesstrax/brokli/pkg/checker"
	"github.com/endlesstrax/brokli/pkg/fetcher"
	"github.com/endlesstrax/brokli/pkg/link"
	"github.com/endlesstrax/brokli/pkg/parser"
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

var verbose bool

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkUrlCmd)
	checkCmd.AddCommand(checkSitemapCmd)

	// Add verbose flag to both subcommands
	checkUrlCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show all links, not just broken ones")
	checkSitemapCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show all URLs, not just broken ones")
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "TODO: Add a short description here",
	Long:  `TODO: Add a longer description here`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var checkUrlCmd = &cobra.Command{
	Use:   "url",
	Short: "Check a URL",
	Long:  `Check a URL`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if URL argument is provided
		if len(args) < 1 {
			fmt.Println("Error: URL argument is required")
			_ = cmd.Help()
			return
		}

		// Take the first argument and use it as the URL
		urlStr := args[0]

		// Parse the URL to get a base URL for resolution
		baseUrl, err := url.Parse(urlStr)
		if err != nil {
			fmt.Printf("Error parsing URL: %v\n", err)
			return
		}

		// Get HTML content
		htmlBytes, err := fetcher.GetHTML(urlStr)
		if err != nil {
			fmt.Printf("Error fetching HTML: %v\n", err)
			return
		}

		// Parse HTML into document tree
		doc, err := parser.ParseHTML(htmlBytes)
		if err != nil {
			fmt.Printf("Error parsing HTML: %v\n", err)
			return
		}

		// Get page results with all links
		results := parser.GetPageResults(doc, *baseUrl)
		fmt.Printf("Found %d links\n", len(results.Links))

		// Check HTTP status for all links
		linkPointers := make([]*link.AnchorTag, len(results.Links))
		for i := range results.Links {
			linkPointers[i] = &results.Links[i]
		}

		ctx := context.Background()
		config := checker.DefaultConfig()

		// Add progress callback
		config.ProgressCallback = func(checked, total int) {
			fmt.Printf("\rChecking links... %d/%d", checked, total)
		}

		err = checker.CheckAnchorTags(ctx, linkPointers, config)
		fmt.Println() // New line after progress

		if err != nil {
			fmt.Printf("Warning: Some links could not be checked: %v\n", err)
		}

		// Count broken links
		brokenCount := 0
		for _, linkPtr := range linkPointers {
			if isBrokenLink(linkPtr.Status) {
				brokenCount++
			}
		}

		// Display results based on verbose flag
		if verbose {
			// Show all links in verbose mode
			fmt.Println("\nAll Links:")
			for i, linkPtr := range linkPointers {
				statusIcon := getStatusIcon(linkPtr.Status)
				coloredStatus := getColoredStatus(linkPtr.Status)
				fmt.Printf("%s %d. %s %s -> %s\n", statusIcon, i+1, coloredStatus, linkPtr.Text, linkPtr.AbsoluteUrl.String())
			}
			// Display summary
			fmt.Printf("\n")
			if brokenCount > 0 {
				color.Red("Summary: %d broken links found out of %d total", brokenCount, len(linkPointers))
			} else {
				color.Green("Summary: All %d links are working", len(linkPointers))
			}
			fmt.Println()
		} else {
			// Show only broken links by default
			if brokenCount > 0 {
				fmt.Println("\nBroken Links:")
				count := 1
				for _, linkPtr := range linkPointers {
					if isBrokenLink(linkPtr.Status) {
						statusIcon := getStatusIcon(linkPtr.Status)
						coloredStatus := getColoredStatus(linkPtr.Status)
						fmt.Printf("%s %d. %s %s -> %s\n", statusIcon, count, coloredStatus, linkPtr.Text, linkPtr.AbsoluteUrl.String())
						count++
					}
				}
				// Display summary for broken links
				fmt.Printf("\n")
				color.Red("Summary: %d broken links found out of %d total", brokenCount, len(linkPointers))
				fmt.Println()
			} else {
				color.Green("\n✓ All links are working!")
				fmt.Println()
			}
		}
	},
}

var checkSitemapCmd = &cobra.Command{
	Use:   "sitemap",
	Short: "Check a sitemap",
	Long:  `Check a sitemap`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if sitemap URL argument is provided
		if len(args) < 1 {
			fmt.Println("Error: sitemap URL argument is required")
			_ = cmd.Help()
			return
		}

		// Take the first argument and use it as the sitemap URL
		sitemapUrl := args[0]

		// Get sitemap XML content
		xmlBytes, err := fetcher.GetHTML(sitemapUrl) // Reuse GetHTML as it fetches any URL content
		if err != nil {
			fmt.Printf("Error fetching sitemap: %v\n", err)
			return
		}

		// Parse sitemap XML
		sitemap, err := parser.ParseSitemap(xmlBytes)
		if err != nil {
			fmt.Printf("Error parsing sitemap: %v\n", err)
			return
		}

		// Get sitemap results
		results := parser.GetSitemapResults(sitemap, sitemapUrl)
		fmt.Printf("Found %d URLs in sitemap\n", len(results.Urls))

		// Check HTTP status for all URLs
		urlPointers := make([]*link.SitemapUrl, len(results.Urls))
		for i := range results.Urls {
			urlPointers[i] = &results.Urls[i]
		}

		ctx := context.Background()
		config := checker.DefaultConfig()

		// Add progress callback
		config.ProgressCallback = func(checked, total int) {
			fmt.Printf("\rChecking URLs... %d/%d", checked, total)
		}

		err = checker.CheckSitemapUrls(ctx, urlPointers, config)
		fmt.Println() // New line after progress

		if err != nil {
			fmt.Printf("Warning: Some URLs could not be checked: %v\n", err)
		}

		// Count broken URLs
		brokenCount := 0
		for _, urlPtr := range urlPointers {
			if isBrokenLink(urlPtr.Status) {
				brokenCount++
			}
		}

		// Display results based on verbose flag
		if verbose {
			// Show all URLs in verbose mode
			fmt.Println("\nAll URLs:")
			for i, urlPtr := range urlPointers {
				statusIcon := getStatusIcon(urlPtr.Status)
				coloredStatus := getColoredStatus(urlPtr.Status)
				fmt.Printf("%s %d. %s %s", statusIcon, i+1, coloredStatus, urlPtr.AbsoluteUrl.String())
				if urlPtr.LastMod != "" {
					fmt.Printf(" (modified: %s)", color.CyanString(urlPtr.LastMod))
				}
				if urlPtr.Priority != "" {
					fmt.Printf(" (priority: %s)", color.YellowString(urlPtr.Priority))
				}
				fmt.Println()
			}
			// Display summary
			fmt.Printf("\n")
			if brokenCount > 0 {
				color.Red("Summary: %d broken URLs found out of %d total", brokenCount, len(urlPointers))
			} else {
				color.Green("Summary: All %d URLs are working", len(urlPointers))
			}
			fmt.Println()
		} else {
			// Show only broken URLs by default
			if brokenCount > 0 {
				fmt.Println("\nBroken URLs:")
				count := 1
				for _, urlPtr := range urlPointers {
					if isBrokenLink(urlPtr.Status) {
						statusIcon := getStatusIcon(urlPtr.Status)
						coloredStatus := getColoredStatus(urlPtr.Status)
						fmt.Printf("%s %d. %s %s", statusIcon, count, coloredStatus, urlPtr.AbsoluteUrl.String())
						if urlPtr.LastMod != "" {
							fmt.Printf(" (modified: %s)", color.CyanString(urlPtr.LastMod))
						}
						if urlPtr.Priority != "" {
							fmt.Printf(" (priority: %s)", color.YellowString(urlPtr.Priority))
						}
						fmt.Println()
						count++
					}
				}
				// Display summary for broken URLs
				fmt.Printf("\n")
				color.Red("Summary: %d broken URLs found out of %d total", brokenCount, len(urlPointers))
				fmt.Println()
			} else {
				color.Green("\n✓ All URLs are working!")
				fmt.Println()
			}
		}
	},
}
