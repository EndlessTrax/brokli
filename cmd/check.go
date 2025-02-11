package cmd

import (
	"fmt"

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
		// Take the first argument and use it as the URL
		r := p.FindAllLinks(p.ParseHTML(p.GetHTML(args[0])))
		fmt.Println(r)
	},
}

var checkUrlCmd = &cobra.Command{
	Use:   "url",
	Short: "Check a URL",
	Long:  `Check a URL`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Check URL command ran")
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