package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/borghives/go-cmd-tool/shared"
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
		creator, _ := cmd.Flags().GetStringSlice("creator")

		if password == "" {
			log.Fatalf("Password is required")
		}

		fmt.Printf("Action: Creating MongoDB admin user '%s'...\n", name)
		client := shared.MustConnectAdminDbClient(&config, false)
		defer client.Disconnect(context.Background())

		newPassword, err := shared.ParseSecretSourceString(password)
		if newPassword == "" {
			log.Fatalf("Failed to extract password: %v", err)
		}

		err = shared.UpsertDbUser(client, name, newPassword, nil, nil, creator, true)
		if err != nil {
			log.Fatalf("Failed to set admin: %v", err)
		}
	},
}

// Define the "list" action command
var listAdminCmd = &cobra.Command{
	Use:   "list",
	Short: "List MongoDB admin",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Listing MongoDB admin...\n")
		client := shared.MustConnectDbClient(&config)
		defer client.Disconnect(context.Background())

		users, err := shared.QueryDbUser(client)
		if err != nil {
			log.Fatalf("Failed to list users: %v", err)
		}

		printUserInfo(users, true)
	},
}

var creator []string

func init() {
	// Add the action to the context
	adminCmd.AddCommand(setAdminCmd)
	adminCmd.AddCommand(listAdminCmd)

	// Define persistent flags
	adminCmd.PersistentFlags().StringP("name", "n", "siteadmin", "Database admin username")

	// Define flags specifically for the 'set' action
	setAdminCmd.Flags().StringP("password", "p", "", "New admin's password")

	setAdminCmd.Flags().StringSliceVarP(&creator, "creator", "c", []string{}, "List of databases the admin can create db and indexes")
}
