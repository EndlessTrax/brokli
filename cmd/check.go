package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/endlesstrax/brokli/pkg/checker"
	"github.com/endlesstrax/brokli/pkg/fetcher"
	"github.com/endlesstrax/brokli/pkg/link"
	"github.com/endlesstrax/brokli/pkg/output"
	"github.com/endlesstrax/brokli/pkg/parser"
)

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkUrlCmd)
	checkCmd.AddCommand(checkSitemapCmd)

	// Add output-format flag to both subcommands
	checkUrlCmd.Flags().StringP("output-format", "o", output.FormatNameDefault, "Output format: default, verbose, or github")
	checkSitemapCmd.Flags().StringP("output-format", "o", output.FormatNameDefault, "Output format: default, verbose, or github")

	// Keep verbose flag for backward compatibility (deprecated)
	checkUrlCmd.Flags().BoolP("verbose", "v", false, "Show all links, not just broken ones (deprecated: use --output-format=verbose)")
	checkSitemapCmd.Flags().BoolP("verbose", "v", false, "Show all URLs, not just broken ones (deprecated: use --output-format=verbose)")
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check links on a URL or sitemap",
	Long:  `Check all links on a webpage or sitemap for broken links. Validates HTTP status codes and displays results with color-coded output.`,
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

		// Determine output format
		formatStr, _ := cmd.Flags().GetString("output-format")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Handle backward compatibility with --verbose flag
		if verbose && formatStr == output.FormatNameDefault {
			formatStr = output.FormatNameVerbose
		}

		// Get the appropriate formatter
		var format output.Format
		switch formatStr {
		case output.FormatNameGitHub:
			format = output.FormatGitHub
		case output.FormatNameVerbose:
			format = output.FormatVerbose
		default:
			format = output.FormatDefault
		}

		formatter := output.NewFormatter(format)

		// Format and display results
		if err := formatter.FormatPageResults(os.Stdout, &results, linkPointers); err != nil {
			fmt.Printf("Error formatting output: %v\n", err)
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

		// Determine output format
		formatStr, _ := cmd.Flags().GetString("output-format")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Handle backward compatibility with --verbose flag
		if verbose && formatStr == output.FormatNameDefault {
			formatStr = output.FormatNameVerbose
		}

		// Get the appropriate formatter
		var format output.Format
		switch formatStr {
		case output.FormatNameGitHub:
			format = output.FormatGitHub
		case output.FormatNameVerbose:
			format = output.FormatVerbose
		default:
			format = output.FormatDefault
		}

		formatter := output.NewFormatter(format)

		// Format and display results
		if err := formatter.FormatSitemapResults(os.Stdout, &results, urlPointers); err != nil {
			fmt.Printf("Error formatting output: %v\n", err)
		}
	},
}
