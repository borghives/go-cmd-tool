package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
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
			Parent: GetProjectParents(cmd),
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

func IsSecretStale(ctx context.Context, client *secretmanager.Client, parent string, secretName string, ttlHours int) bool {
	name := fmt.Sprintf("%s/secrets/%s/versions/latest", parent, secretName)
	req := &secretmanagerpb.GetSecretVersionRequest{
		Name: name,
	}
	secret, err := client.GetSecretVersion(ctx, req)
	if err != nil {
		fmt.Printf("Failed to get secret: %s %v\n", err)
		return true
	}
	return secret.CreateTime.AsTime().Before(time.Now().Add(-time.Hour * time.Duration(ttlHours)))
}

func init() {
	rootCmd.AddCommand(listCmd)
}
