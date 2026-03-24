package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// Define the "admin" context command
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage MongoDB admin user",
}

// Define the "create" action command
var setAdminCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a new MongoDB admin",
	Run: func(cmd *cobra.Command, args []string) {
		password, _ := cmd.Flags().GetString("password")
		name, _ := cmd.Flags().GetString("name")

		fmt.Printf("Action: Creating MongoDB admin user '%s'...\n", name)
		client := GetDbClient(cmd)
		defer client.Disconnect(context.Background())

		err := UpsertDbUser(client, name, password, nil, nil, true)
		if err != nil {
			log.Fatalf("Failed to set admin: %v", err)
		}
	},
}

// Define the "remove" action command
var removeAdminCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a MongoDB admin",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			log.Fatalf("Name is required")
		}

		fmt.Printf("Action: Remove MongoDB admin '%s'...\n", name)
		client := GetDbClient(cmd)
		defer client.Disconnect(context.Background())

		err := DeleteDbUser(client, name)
		if err != nil {
			log.Fatalf("Failed to remove admin: %v", err)
		}
	},
}

// Define the "list" action command
var listAdminCmd = &cobra.Command{
	Use:   "list",
	Short: "List MongoDB admin",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Listing MongoDB admin...\n")
		client := GetDbClient(cmd)
		defer client.Disconnect(context.Background())

		users, err := QueryDbUser(client)
		if err != nil {
			log.Fatalf("Failed to list users: %v", err)
		}

		printUserInfo(users, true)
	},
}

func init() {
	// Add the action to the context
	adminCmd.AddCommand(setAdminCmd)
	adminCmd.AddCommand(listAdminCmd)
	adminCmd.AddCommand(removeAdminCmd)

	// Add the context to the root dbenv command
	rootCmd.AddCommand(adminCmd)

	// Set Client Flags
	SetClientFlags(adminCmd)

	// Define persistent flags
	adminCmd.PersistentFlags().StringP("name", "n", "siteadmin", "Database admin username")

	// Define flags specifically for the 'set' action
	setAdminCmd.Flags().StringP("password", "p", "", "New admin's password")
}
