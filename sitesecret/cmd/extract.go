package cmd

import (
	"fmt"
	"log"

	"github.com/borghives/kosmos-go"
	"github.com/spf13/cobra"
)

// Define the "list" context command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract secrets",
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("source")
		if source == "" {
			log.Fatalf("Secret source is required")
		}

		secret, err := kosmos.CollapseSecret(source)
		if err != nil {
			log.Fatalf("failed to collapse secret: %v", err)
		}
		fmt.Print(secret)

	},
}

func init() {
	extractCmd.Flags().StringP("source", "s", "", "Secret source string")
}
