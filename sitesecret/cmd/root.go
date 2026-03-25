package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sitesecret",
	Short: "A CLI tool to manage secret for a site",
}

// Execute is called by main.main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("namespace", "s", "", "Namespace prefix")
	rootCmd.PersistentFlags().StringP("project", "p", "", "Project ID")
}
