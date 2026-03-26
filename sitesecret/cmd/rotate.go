package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/borghives/go-cmd-tool/shared"
	"github.com/spf13/cobra"
)

// Define the "list" context command
var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate a secret",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Rotating a secret...\n")
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			log.Fatalf("Secret name is required")
		}

		namespace, _ := cmd.Flags().GetString("namespace")
		if namespace != "" {
			namespace = namespace + "_"

		}

		secretName := namespace + name
		fmt.Printf("Secret name: %s\n", secretName)

		payload := shared.GeneratePayload(cmd)

		// 1. Build the request to list secrets
		ctx := context.Background()
		client := shared.MustGetSecretClient(ctx)
		defer client.Close()

		ttl, _ := cmd.Flags().GetInt("ttl")

		projectParent := shared.GetProjectParents()

		if shared.IsSecretStale(ctx, client, projectParent, secretName, ttl) {
			fmt.Println("Generating random payload for secret.")
			payload = shared.GenerateRandomString(32)
			shared.CreateSecret(ctx, client, projectParent, secretName)
			shared.AddSecretVersion(ctx, client, projectParent, secretName, payload)
		} else {
			fmt.Println("Secret is fresh. No rotation needed.")
		}
	},
}

func init() {
	rootCmd.AddCommand(rotateCmd)

	rotateCmd.Flags().StringP("name", "n", "", "Secret name")
	rotateCmd.Flags().IntP("ttl", "", 24, "Secret time to live in hours.  If the secret is older than this value, it will be rotated.")

}
