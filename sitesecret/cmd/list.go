package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/borghives/go-cmd-tool/shared"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

// Define the "list" context command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Listing secrets...\n")

		// 1. Build the request to list secrets
		req := &secretmanagerpb.ListSecretsRequest{
			Parent: shared.GetProjectParents(),
		}

		// 2. Create the Secret Manager client
		ctx := context.Background()
		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			log.Fatalf("failed to setup client: %v", err)
		}
		defer client.Close()

		namespace, _ := cmd.Flags().GetString("namespace")
		// 3. Iterate over the secrets
		it := client.ListSecrets(ctx, req)
		for {
			secret, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("failed to list secrets: %v", err)
			}

			secretParts := strings.Split(secret.Name, "/")
			secretName := secretParts[3]

			if namespace != "" && !strings.HasPrefix(secretName, namespace+"_") {
				continue
			}

			fmt.Printf("- %s\n", secretName)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
