package cmd

import (
	"os"

	"github.com/borghives/kosmos-go/ether"
	"github.com/spf13/cobra"
)

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
	rootCmd.PersistentFlags().StringP("uri", "u", "", "MongoDB connection URI")
	rootCmd.AddCommand(adminCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(rsCmd)

	cobra.OnInitialize(func() {
		ether.CollapseConstants().MergeFromFile("tool.env").MergeFromCmd(rootCmd)
		ether.ColapseObserverConstants().MergeFromFile("tool.env").MergeFromCmd(rootCmd)
	})
}
