package cmd

import (
	"fmt"
	"log"

	"github.com/borghives/kosmos-go"
	"github.com/spf13/cobra"
)

// Define the "list" context command
var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate a secret",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Rotating a secret...\n")
		secretName, _ := cmd.Flags().GetString("name")
		if secretName == "" {
			log.Fatalf("Secret name is required")
		}

		fmt.Printf("Secret name: %s\n", secretName)

		manager := kosmos.SummonSecretManager()

		ttl, _ := cmd.Flags().GetInt("ttl")
		if manager.IsSecretStale(secretName, ttl) {
			fmt.Println("Generating random payload for secret.")
			payload := GenerateRandomString(32)
			manager.CreateSecret(secretName)
			manager.AddSecretVersion(secretName, payload)
		} else {
			fmt.Println("Secret is fresh. No rotation needed.")
		}
	},
}

func init() {
	rotateCmd.Flags().StringP("name", "n", "", "Secret name")
	rotateCmd.Flags().IntP("ttl", "", 24, "Secret time to live in hours.  If the secret is older than this value, it will be rotated.")

}
