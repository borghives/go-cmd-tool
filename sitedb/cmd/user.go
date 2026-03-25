package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/borghives/go-cmd-tool/shared"
	"github.com/spf13/cobra"
)

// Define the "user" context command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage MongoDB users",
}

// Define the "set" action command
var setUserCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a MongoDB user",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		password, _ := cmd.Flags().GetString("password")

		if name == "" || password == "" {
			log.Fatalf("Name and password are required")
		}

		readDb, _ := cmd.Flags().GetStringSlice("read")
		readWriteDb, _ := cmd.Flags().GetStringSlice("write")

		fmt.Printf("Action: Set MongoDB user '%s'...\n", name)
		fmt.Printf("Read permission: %v\n", readDb)
		fmt.Printf("ReadWrite permission: %v\n", readWriteDb)
		client := shared.MustGetDbClient(config.OverrideFromCmd(cmd))
		defer client.Disconnect(context.Background())

		newPassword, err := shared.ParseSecretHolderString(config.OverrideFromCmd(cmd), password)
		if err != nil {
			log.Fatalf("Failed to parse password: %v", err)
		}

		err = shared.UpsertDbUser(client, name, newPassword, readDb, readWriteDb, false)
		if err != nil {
			log.Fatalf("Failed to set user: %v", err)
		}
	},
}

// Define the "remove" action command
var removeUserCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a MongoDB user",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			log.Fatalf("Name is required")
		}

		fmt.Printf("Action: Remove MongoDB user '%s'...\n", name)
		client := shared.MustGetDbClient(config.OverrideFromCmd(cmd))
		defer client.Disconnect(context.Background())

		err := shared.DeleteDbUser(client, name)
		if err != nil {
			log.Fatalf("Failed to remove user: %v", err)
		}
	},
}

// Define the "list" action command
var listUserCmd = &cobra.Command{
	Use:   "list",
	Short: "List MongoDB users",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Listing MongoDB users...\n")
		client := shared.MustGetDbClient(config.OverrideFromCmd(cmd))
		defer client.Disconnect(context.Background())

		users, err := shared.QueryDbUser(client)
		if err != nil {
			log.Fatalf("Failed to list users: %v", err)
		}

		printUserInfo(users, false)
	},
}

func printUserInfo(res *shared.UsersInfoResponse, filterAdmin bool) {
	if res == nil {
		log.Fatalf("Empty User Info")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	if filterAdmin {
		fmt.Fprintf(w, "ADMIN\tROLES\n")
	} else {
		fmt.Fprintf(w, "USER\tROLES\n")
	}

	fmt.Fprintln(w, "--------\t---------")

	for _, u := range res.Users {
		// Format roles as a comma-separated string: "role1 (db), role2 (db)"
		var roles []string
		for _, r := range u.Roles {
			roles = append(roles, fmt.Sprintf("%s (%s)", r.Role, r.DB))
		}

		if filterAdmin {
			if !slices.Contains(roles, "userAdminAnyDatabase (admin)") {
				continue
			}
		}

		roleList := strings.Join(roles, ", ")
		fmt.Fprintf(w, "%s\t%s\n", u.User, roleList)
	}
	w.Flush()
}

var readDb []string
var readWriteDb []string

func init() {
	// Add the action to the context
	userCmd.AddCommand(setUserCmd)
	userCmd.AddCommand(listUserCmd)
	userCmd.AddCommand(removeUserCmd)

	// Add the context to the root dbenv command
	rootCmd.AddCommand(userCmd)

	// Set Client Flags
	shared.SetDbClientFlags(userCmd)

	// Define persistent flags
	userCmd.PersistentFlags().StringP("name", "n", "", "Database username")

	// Define flags specifically for the 'create' action
	setUserCmd.Flags().StringP("password", "p", "", "Password for the new user")
	setUserCmd.Flags().StringSliceVarP(&readDb, "read", "r", []string{}, "List of read database (comma-separated or multiple flags)")
	setUserCmd.Flags().StringSliceVarP(&readWriteDb, "write", "w", []string{}, "List of readWrite database (comma-separated or multiple flags)")
}
