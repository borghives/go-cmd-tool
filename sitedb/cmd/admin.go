package cmd

import (
	"fmt"
	"log"

	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/observation"

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

		observer := kosmos.SummonObserverFor(observation.PurposeAffinityAdmin)
		defer observer.Close()

		var err error
		if kosmos.IsSecretSource(password) {
			password, err = kosmos.CollapseSecret(password)
			if err != nil {
				log.Fatalf("Failed to extract password: %v", err)
			}
		}

		responsibility := observation.MemberResponsibility{
			CreatorDbs: creator,
			IsAdmin:    true,
		}

		err = observer.UpdateMember(name, password, responsibility, true)
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
		observer := kosmos.SummonObserverFor(observation.PurposeAffinityAdmin)
		defer observer.Close()

		users, err := observer.ListMembers()
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
