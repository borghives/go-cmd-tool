package shared

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

func GetDbClient(cmd *cobra.Command) *mongo.Client {
	uriFmt, _ := cmd.Flags().GetString("uri")
	if uriFmt == "" {
		uriFmt = os.Getenv("MONGODB_URI")
		if uriFmt == "" {
			fmt.Printf("Using default MongoDB URI: mongodb://127.0.0.1:27017/\n")
			uriFmt = "mongodb://127.0.0.1:27017/"
		}
	}
	uri := os.ExpandEnv(uriFmt)

	//translate uri
	var err error
	uri, err = TranslateMongoURIPassword(uri)

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	fmt.Printf("Pinging...")
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping client: %v", err)
	}
	fmt.Printf("Connected!\n")
	return client
}

func SetDbClientFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("uri", "u", "", "MongoDB connection URI")
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
	newPassword = ParseHolderString(newPassword)

	if newPassword == "" {
		return fmt.Errorf("Cannot set empty password\n")
	}

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
	newPassword = ParseHolderString(newPassword)

	if newPassword == "" {
		return fmt.Errorf("Cannot set empty password\n")
	}

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
