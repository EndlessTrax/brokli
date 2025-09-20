package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	p "github.com/endlesstrax/brokli/pkg/parse"
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
		cmd.Help()
	},
}

var checkUrlCmd = &cobra.Command{
	Use:   "url",
	Short: "Check a URL",
	Long:  `Check a URL`,
	Run: func(cmd *cobra.Command, args []string) {
		// Take the first argument and use it as the URL
		urlStr := args[0]

		// Parse the URL to get a base URL for resolution
		baseUrl, err := url.Parse(urlStr)
		if err != nil {
			fmt.Printf("Error parsing URL: %v\n", err)
			return
		}

		// Get HTML content
		htmlBytes, err := p.GetHTML(urlStr)
		if err != nil {
			fmt.Printf("Error fetching HTML: %v\n", err)
			return
		}

		// Parse HTML into document tree
		doc, err := p.ParseHTML(htmlBytes)
		if err != nil {
			fmt.Printf("Error parsing HTML: %v\n", err)
			return
		}

		// Get page results with all links
		results := p.GetPageResults(doc, *baseUrl)
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
		fmt.Println("Check sitemap command ran")
	},
}
