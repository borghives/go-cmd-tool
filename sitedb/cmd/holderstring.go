package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

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

func GetProjectParents() string {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		fmt.Println("Project flag and environment PROJECT_ID is not set")
		return ""
	}

	return fmt.Sprintf("projects/%s", projectID)
}

func ExtractSecret(name string, version string) string {
	//get the secret from the secret manager

	projectParents := GetProjectParents()
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
