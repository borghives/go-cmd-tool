package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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

func SetClientFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("uri", "u", "", "MongoDB connection URI")
}
