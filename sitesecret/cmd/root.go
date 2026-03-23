package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sitesecret",
	Short: "A CLI tool to manage secret for a site",
}

func GetProjectParents(cmd *cobra.Command) string {
	projectID, _ := cmd.Flags().GetString("project")
	if projectID == "" {
		projectID = os.Getenv("PROJECT_ID")
		if projectID == "" {
			log.Fatalf("Project flag and environment PROJECT_ID is not set")
		}
	}
	return fmt.Sprintf("projects/%s", projectID)
}

func MustGetSecretClient(ctx context.Context) *secretmanager.Client {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to setup client: %v", err)
	}
	return client
}

// Execute is called by main.main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("namespace", "s", "", "Namespace prefix")
	rootCmd.PersistentFlags().StringP("project", "p", "", "Project ID")
}
