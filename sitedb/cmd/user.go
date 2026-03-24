package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoRole struct {
	Role string `bson:"role"`
	DB   string `bson:"db"`
}

type MongoUser struct {
	User  string      `bson:"user"`
	Roles []MongoRole `bson:"roles"`
}

type UsersInfoResponse struct {
	Users []MongoUser `bson:"users"`
}

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
		client := GetDbClient(cmd)
		defer client.Disconnect(context.Background())

		err := UpsertDbUser(client, name, password, readDb, readWriteDb, false)
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
		client := GetDbClient(cmd)
		defer client.Disconnect(context.Background())

		err := DeleteDbUser(client, name)
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
		client := GetDbClient(cmd)
		defer client.Disconnect(context.Background())

		users, err := QueryDbUser(client)
		if err != nil {
			log.Fatalf("Failed to list users: %v", err)
		}

		printUserInfo(users, false)
	},
}

func QueryDbUser(client *mongo.Client) (*UsersInfoResponse, error) {
	usersInfoCmd := bson.D{
		{Key: "usersInfo", Value: 1},
	}

	var result UsersInfoResponse
	err := client.Database("admin").RunCommand(context.Background(), usersInfoCmd).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func printUserInfo(res *UsersInfoResponse, filterAdmin bool) {
	if res == nil {
		log.Fatalf("Empty User Info")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Print Header
	fmt.Fprintln(w, "USER\tROLES")
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

func translateRole(readDb []string, readWriteDb []string, isAdmin bool) bson.A {
	roles := bson.A{}
	for _, db := range readDb {
		roles = append(roles, bson.M{"role": "read", "db": db})
	}
	for _, db := range readWriteDb {
		roles = append(roles, bson.M{"role": "readWrite", "db": db})
	}
	if isAdmin {
		roles = append(roles, bson.M{"role": "userAdminAnyDatabase", "db": "admin"})
	}
	return roles
}

func CreateDbUser(client *mongo.Client, username string, newPassword string, readDb []string, readWriteDb []string, isAdmin bool) error {
	createUserCmd := bson.D{
		{Key: "createUser", Value: username},
		{Key: "pwd", Value: newPassword},
		{Key: "roles", Value: translateRole(readDb, readWriteDb, isAdmin)},
	}

	var result bson.M
	err := client.Database("admin").RunCommand(context.Background(), createUserCmd).Decode(&result)
	if err != nil {
		return err
	}

	if isAdmin {
		fmt.Printf("Successfully created admin: %s\n", username)
	} else {
		fmt.Printf("Successfully created user: %s\n", username)
	}

	return nil
}

func UpdateDbUser(client *mongo.Client, username string, newPassword string, readDb []string, readWriteDb []string, isAdmin bool) error {
	roles := translateRole(readDb, readWriteDb, isAdmin)
	fmt.Printf("Roles: %v\n", roles)
	updateUserCmd := bson.D{
		{Key: "updateUser", Value: username},
		{Key: "pwd", Value: newPassword},
		{Key: "roles", Value: roles},
	}

	var result bson.M
	err := client.Database("admin").RunCommand(context.Background(), updateUserCmd).Decode(&result)
	if err != nil {
		return err
	}

	if isAdmin {
		fmt.Printf("Successfully updated admin: %s\n", username)
	} else {
		fmt.Printf("Successfully updated user: %s\n", username)
	}

	return nil
}

func UpsertDbUser(client *mongo.Client, username string, newPassword string, readDb []string, readWriteDb []string, isAdmin bool) error {
	err := UpdateDbUser(client, username, newPassword, readDb, readWriteDb, isAdmin)
	if err != nil {
		return CreateDbUser(client, username, newPassword, readDb, readWriteDb, isAdmin)
	}
	return nil
}

func DeleteDbUser(client *mongo.Client, username string) error {
	dropUserCmd := bson.D{
		{Key: "dropUser", Value: username},
	}

	var result bson.M
	err := client.Database("admin").RunCommand(context.Background(), dropUserCmd).Decode(&result)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully removed user: %s\n", username)

	return nil
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
	SetClientFlags(userCmd)

	// Define persistent flags
	userCmd.PersistentFlags().StringP("name", "n", "", "Database username")

	// Define flags specifically for the 'create' action
	setUserCmd.Flags().StringP("password", "p", "", "Password for the new user")
	setUserCmd.Flags().StringSliceVarP(&readDb, "read", "r", []string{}, "List of read database (comma-separated or multiple flags)")
	setUserCmd.Flags().StringSliceVarP(&readWriteDb, "write", "w", []string{}, "List of readWrite database (comma-separated or multiple flags)")
}
