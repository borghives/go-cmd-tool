package cmd

import (
	"fmt"
	"os"

	"github.com/borghives/go-cmd-tool/shared"
	"github.com/spf13/cobra"
)

var config shared.SiteConfig

var rootCmd = &cobra.Command{
	Use:   "sitedb",
	Short: "A CLI tool to manage MongoDB environments for a site",
}

// Execute is called by main.main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Set Client Flags
	shared.SetDbClientFlags(rootCmd)

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
