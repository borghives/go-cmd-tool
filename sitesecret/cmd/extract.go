package cmd

import (
	"fmt"
	"log"

	"github.com/borghives/go-cmd-tool/shared"
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

		secret, err := shared.ParseSecretSourceString(source)
		if err == nil {
			fmt.Print(secret)
			return
		}

		if secret == "" && err != nil {
			log.Fatalf("Failed to parse secret source string: %v", err)
		}

		secret, err = shared.ParseSecretVersionName(source)
		if err != nil {
			log.Fatalf("Failed to parse secret version name: %v", err)
		}

		fmt.Print(secret)

	},
}

func init() {
	extractCmd.Flags().StringP("source", "s", "", "Secret source string")
}
