package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "brokli",
	Version: "0.1",
	Short: "TODO: Add a short description here",
	Long: `TODO: Add a longer description here`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		fmt.Println("root command ran")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Brokli",
	Long:  `All software has versions. This is Brokli's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Brokli v0.1")
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