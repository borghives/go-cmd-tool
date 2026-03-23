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

func init() {
	// Add the action to the context
	adminCmd.AddCommand(setAdminCmd)

	// Add the context to the root dbenv command
	rootCmd.AddCommand(adminCmd)

	// Set Client Flags
	SetClientFlags(adminCmd)

	// Define flags specifically for the 'set' action
	setAdminCmd.Flags().StringP("name", "n", "siteadmin", "Name for the new admin.")
	setAdminCmd.Flags().StringP("password", "p", "", "New admin's password")

}
