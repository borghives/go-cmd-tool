package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/borghives/go-cmd-tool/shared"
	"github.com/spf13/cobra"
)

// Define the "list" context command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new secret",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Creating a new secret...\n")
		secretName, _ := cmd.Flags().GetString("name")
		if secretName == "" {
			log.Fatalf("Secret name is required")
		}

		fmt.Printf("Secret name: %s\n", secretName)

		payload := shared.GeneratePayload(cmd)

		// 1. Build the request to list secrets
		ctx := context.Background()
		client := shared.MustGetSecretClient(ctx)
		defer client.Close()

		projectParent := shared.GetProjectParents()

		//ignore error
		shared.CreateSecret(ctx, client, projectParent, secretName)
		shared.AddSecretVersion(ctx, client, projectParent, secretName, payload)
	},
}

func init() {
	newCmd.Flags().StringP("name", "n", "", "Secret name")
	newCmd.Flags().StringP("payload", "", "", "Secret payload")

}
