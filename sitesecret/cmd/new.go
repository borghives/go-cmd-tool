package cmd

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/spf13/cobra"
)

// Define the "list" context command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new secret",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Creating a new secret...\n")
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

		payload := GeneratePayload(cmd)

		// 1. Build the request to list secrets
		ctx := context.Background()
		client := MustGetSecretClient(ctx)
		defer client.Close()

		//ignore error
		CreateSecret(ctx, client, GetProjectParents(cmd), secretName)
		AddSecretVersion(ctx, client, GetProjectParents(cmd), secretName, payload)
	},
}

func AddSecretVersion(ctx context.Context, client *secretmanager.Client, parent string, secretName string, payload string) error {
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: parent + "/secrets/" + secretName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(payload),
		},
	}

	version, err := client.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		log.Fatalf("failed to add secret version: %v", err)
	}
	fmt.Printf("Successfully added secret version: %s\n", version.Name)
	return nil
}

func GeneratePayload(cmd *cobra.Command) string {
	payload, _ := cmd.Flags().GetString("payload")
	if payload == "" {
		fmt.Println("Generating random string for payload.")
		payload = GenerateRandomString(32)
	}
	return payload
}

func GenerateRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, n)

	// Read random data from the OS into the byte slice
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}

	for i, b := range bytes {
		// Use modulo to map the random byte to our charset
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}

func CreateSecret(ctx context.Context, client *secretmanager.Client, parent string, secretName string) error {
	// 1. Create the Secret (Metadata container)
	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   parent,
		SecretId: secretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	_, err := client.CreateSecret(ctx, createSecretReq)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully created new secret: %s\n", secretName)
	return nil
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringP("name", "n", "", "Secret name")
	newCmd.Flags().StringP("payload", "", "", "Secret payload")

}
