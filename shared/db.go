package shared

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/net/proxy"
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

type mongoDialerWrapper struct {
	dialer proxy.Dialer
}

func (m *mongoDialerWrapper) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if cd, ok := m.dialer.(interface {
		DialContext(context.Context, string, string) (net.Conn, error)
	}); ok {
		return cd.DialContext(ctx, network, addr)
	}
	return m.dialer.Dial(network, addr)
}

func MustGetDbClient(cfg *SiteConfig) *mongo.Client {
	uri := ""
	proxyAddress := ""

	if cfg != nil {
		uri = cfg.MongoDBAuthUri
		if uri == "" {
			uri = cfg.MongoDBUri
		}
		proxyAddress = cfg.ProxyAddress
	}

	if uri == "" {
		log.Printf("Using default MongoDB URI: mongodb://127.0.0.1:27017/\n")
		uri = "mongodb://127.0.0.1:27017/"
	}

	uri = os.ExpandEnv(uri)

	//translate uri
	var err error
	uri, err = TranslateMongoURIPassword(cfg, uri)
	if err != nil {
		log.Fatalf("Failed to translate secret from URI: %v", err)
	}

	clientOptions := options.Client().ApplyURI(uri)

	if proxyAddress != "" {
		log.Println("Using proxy: ", proxyAddress)
		proxyUrl, err := url.Parse(proxyAddress)
		if err != nil {
			log.Fatalf("Failed to parse proxy address: %v", err)
		}

		dialer, err := proxy.FromURL(proxyUrl, proxy.Direct)
		if err != nil {
			log.Fatalf("Failed to create dialer: %v", err)
		}

		clientOptions = clientOptions.SetDialer(&mongoDialerWrapper{dialer: dialer})
	}

	var client *mongo.Client
	for i := range 5 {
		client, err = mongo.Connect(clientOptions)
		if err == nil {
			err = client.Ping(context.Background(), nil)
		}
		if err == nil {
			log.Printf("MongoDb Ping Success")
			break
		}
		log.Printf("MongoDb Ping Failed.  Waiting for MongoDB... (attempt %d): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if client == nil {
		log.Fatalf("Failed to connect to MongoDB")
	}

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
	if newPassword == "" {
		return fmt.Errorf("Cannot set empty password\n")
	}

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
