package cmd

import (
	"fmt"
	"log"

	"github.com/borghives/kosmos-go"
	"github.com/spf13/cobra"
)

// Define the "list" context command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Listing secrets...\n")
		secrets, err := kosmos.SummonSecretManager().ListSecrets()
		if err != nil {
			log.Fatalf("failed to list secrets: %v", err)
		}

		for _, secret := range secrets {
			fmt.Printf("- %s\n", secret.Name)
		}
	},
}
