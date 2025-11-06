package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/endlesstrax/brokli/pkg/fetcher"
	"github.com/endlesstrax/brokli/pkg/parser"
)

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
		fmt.Printf("Found %d links:\n", len(results.Links))
		for i, link := range results.Links {
			fmt.Printf("%d. %s -> %s\n", i+1, link.Text, link.AbsoluteUrl.String())
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
		fmt.Printf("Found %d URLs in sitemap:\n", len(results.Urls))
		for i, sitemapUrl := range results.Urls {
			fmt.Printf("%d. %s", i+1, sitemapUrl.AbsoluteUrl.String())
			if sitemapUrl.LastMod != "" {
				fmt.Printf(" (last modified: %s)", sitemapUrl.LastMod)
			}
			if sitemapUrl.Priority != "" {
				fmt.Printf(" (priority: %s)", sitemapUrl.Priority)
			}
			fmt.Println()
		}
	},
}
