package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is the version of Brokli.
var Version = "0.2.1"

var rootCmd = &cobra.Command{
	Use:     "brokli",
	Version: Version,
	Short:   "A fast, concurrent broken link checker for websites and sitemaps",
	Long:    `Brokli is a CLI tool that helps developers validate all links on a page or sitemap during development. It checks HTTP status codes concurrently and displays results with color-coded output, making it easy to spot broken links before deployment.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, show help
		_ = cmd.Help()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Brokli",
	Long:  `All software has versions. This is Brokli's`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print the version number of Brokli
		fmt.Println("Brokli v" + Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
