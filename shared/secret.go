package shared

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/spf13/cobra"
)

func MustGetSecretClient(ctx context.Context) *secretmanager.Client {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to setup client: %v", err)
	}
	return client
}

func IsSecretStale(ctx context.Context, client *secretmanager.Client, parent string, secretName string, ttlHours int) bool {
	name := fmt.Sprintf("%s/secrets/%s/versions/latest", parent, secretName)
	req := &secretmanagerpb.GetSecretVersionRequest{
		Name: name,
	}
	secret, err := client.GetSecretVersion(ctx, req)
	if err != nil {
		fmt.Printf("Failed to get secret: %v\n", err)
		return true
	}
	return secret.CreateTime.AsTime().Before(time.Now().Add(-time.Hour * time.Duration(ttlHours)))
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

func TranslateMongoURIPassword(uri string) (string, error) {
	// 1. Isolate the scheme
	schemeSplit := strings.SplitN(uri, "://", 2)
	if len(schemeSplit) != 2 {
		return "", fmt.Errorf("invalid URI format")
	}
	scheme, remainder := schemeSplit[0], schemeSplit[1]

	// 2. Find the end of the credentials (the LAST '@' before any '/' or '?')
	// This is important because passwords themselves can contain '@' if encoded,
	// but the delimiter between creds and hosts is the final '@'.
	endOfCreds := strings.LastIndex(remainder, "@")
	if endOfCreds == -1 {
		return uri, nil // No credentials found (unauthenticated connection)
	}

	creds := remainder[:endOfCreds]
	hostAndPath := remainder[endOfCreds:] // Includes the '@'

	// 3. Split User and Password
	userPass := strings.SplitN(creds, ":", 2)
	if len(userPass) < 2 {
		return uri, nil // Only user, no password
	}

	user, pass := userPass[0], userPass[1]

	// 4. Translate and Stitch
	newPass := ParseHolderString(pass)
	return fmt.Sprintf("%s://%s:%s%s", scheme, user, newPass, hostAndPath), nil
}

func ParseHolderString(s string) string {
	//parse string "[secret:name:version]"
	//return the secret

	//check if the string is a holder string
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return s
	}

	//remove the brackets
	s = s[1 : len(s)-1]

	//split the string by ":"
	parts := strings.Split(s, ":")

	//check if the string is a holder string
	if len(parts) != 3 {
		return s
	}

	if parts[0] != "secret" {
		return s
	}

	//return the secret
	return ExtractSecret(parts[1], parts[2])
}

func ExtractSecret(name string, version string) string {
	//get the secret from the secret manager

	projectParents := GetProjectParents(nil)
	if projectParents == "" {
		fmt.Println("Failed to find project_id to extract secret from")
		return ""
	}

	// 1. Build the request to get secrets
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/secrets/%s/versions/%s", projectParents, name, version),
	}

	// 2. Create the Secret Manager client
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to setup client: %v", err)
	}
	defer client.Close()

	resp, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		fmt.Printf("failed to extract secret: %v", err)
		return ""
	}

	return string(resp.Payload.Data)
}
