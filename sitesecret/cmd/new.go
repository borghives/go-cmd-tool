package cmd

import (
	"crypto/rand"
	"fmt"
	"log"

	"github.com/borghives/kosmos-go"
	"github.com/spf13/cobra"
)

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

		payload := GeneratePayload(cmd)

		// 1. Build the request to list secrets
		manager := kosmos.SummonSecretManager()

		//ignore error
		manager.CreateSecret(secretName)
		manager.AddSecretVersion(secretName, payload)
	},
}

func init() {
	newCmd.Flags().StringP("name", "n", "", "Secret name")
	newCmd.Flags().StringP("payload", "", "", "Secret payload")

}
