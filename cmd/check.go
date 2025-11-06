package cmd

import (
	"context"
	"fmt"
	"net/url"

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

// isBrokenLink returns true if the status code indicates a broken link
func isBrokenLink(status int) bool {
	// A link is broken if it's unchecked (-1) or has a 4xx/5xx status code
	return status == -1 || status >= 400
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkUrlCmd)
	checkCmd.AddCommand(checkSitemapCmd)
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
		fmt.Println("Checking link status...")
		linkPointers := make([]*link.AnchorTag, len(results.Links))
		for i := range results.Links {
			linkPointers[i] = &results.Links[i]
		}

		ctx := context.Background()
		config := checker.DefaultConfig()
		err = checker.CheckAnchorTags(ctx, linkPointers, config)
		if err != nil {
			fmt.Printf("Warning: Some links could not be checked: %v\n", err)
		}

		// Display results
		fmt.Println("\nResults:")
		brokenCount := 0
		for i, linkPtr := range linkPointers {
			statusIcon := getStatusIcon(linkPtr.Status)
			fmt.Printf("%s %d. [%d] %s -> %s\n", statusIcon, i+1, linkPtr.Status, linkPtr.Text, linkPtr.AbsoluteUrl.String())
			if isBrokenLink(linkPtr.Status) {
				brokenCount++
			}
		}

		fmt.Printf("\nSummary: %d total links, %d broken\n", len(linkPointers), brokenCount)
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
		fmt.Println("Checking URL status...")
		urlPointers := make([]*link.SitemapUrl, len(results.Urls))
		for i := range results.Urls {
			urlPointers[i] = &results.Urls[i]
		}

		ctx := context.Background()
		config := checker.DefaultConfig()
		err = checker.CheckSitemapUrls(ctx, urlPointers, config)
		if err != nil {
			fmt.Printf("Warning: Some URLs could not be checked: %v\n", err)
		}

		// Display results
		fmt.Println("\nResults:")
		brokenCount := 0
		for i, urlPtr := range urlPointers {
			statusIcon := getStatusIcon(urlPtr.Status)
			fmt.Printf("%s %d. [%d] %s", statusIcon, i+1, urlPtr.Status, urlPtr.AbsoluteUrl.String())
			if urlPtr.LastMod != "" {
				fmt.Printf(" (modified: %s)", urlPtr.LastMod)
			}
			if urlPtr.Priority != "" {
				fmt.Printf(" (priority: %s)", urlPtr.Priority)
			}
			fmt.Println()
			if isBrokenLink(urlPtr.Status) {
				brokenCount++
			}
		}

		fmt.Printf("\nSummary: %d total URLs, %d broken\n", len(urlPointers), brokenCount)
	},
}
