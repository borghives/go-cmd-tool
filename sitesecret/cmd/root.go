package cmd

import (
	"fmt"
	"os"

	"github.com/borghives/go-cmd-tool/shared"
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

var config shared.SiteConfig

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(extractCmd)
	rootCmd.AddCommand(rotateCmd)

	rootCmd.PersistentFlags().StringP("project", "p", "", "Project ID")

	cobra.OnInitialize(func() {
		var err error
		config, err = shared.LoadSiteConfig()
		config.MergeFromFile("tool.env")
		config.MergeFromCmd(rootCmd)
		if err != nil {
			fmt.Printf("Failed to load site config: %v\n", err)
			os.Exit(1)
		}
	})
}
