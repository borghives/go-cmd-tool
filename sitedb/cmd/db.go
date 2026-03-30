package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/observation"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gopkg.in/yaml.v3"
)

// IndexKey represents a single field and its direction in a MongoDB index.
type IndexKey struct {
	Field string `yaml:"field"`
	Order int32  `yaml:"order"`
}

// IndexConfig represents the configuration for a single MongoDB index.
type IndexConfig struct {
	Keys   []IndexKey `yaml:"keys"`
	Unique bool       `yaml:"unique"`
}

type TimeseriesInfo struct {
	TimeField string `yaml:"time-field"`
	MetaField string `yaml:"meta-field"`
}

// CollectionConfig represents the configuration for a MongoDB collection and its internal indexes.
type CollectionConfig struct {
	Name           string          `yaml:"name"`
	Indexes        []IndexConfig   `yaml:"indexes"`
	TimeseriesInfo *TimeseriesInfo `yaml:"timeseries-info,omitempty"`
}

// DatabaseConfig represents a MongoDB database and its collections.
type DatabaseConfig struct {
	Name        string             `yaml:"name"`
	Collections []CollectionConfig `yaml:"collections"`
}

type UserConfig struct {
	Name      string   `yaml:"name"`
	SecretSrc string   `yaml:"secret-src"`
	Read      []string `yaml:"read"`
	ReadWrite []string `yaml:"write"`
	Creator   []string `yaml:"creator"`
	Admin     bool     `yaml:"admin"`
}

// DataConfig is the root structure for data initialization YAML file.
type DataConfig struct {
	Databases []DatabaseConfig `yaml:"databases"`
	Users     []UserConfig     `yaml:"users"`
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage MongoDB data, collections, and indexes",
}

var dbDeclareCmd = &cobra.Command{
	Use:   "declare",
	Short: "Declare database, collections, and indexes from a YAML configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			log.Fatalf("Error: --file flag is required")
		}

		fmt.Printf("Action: Declaring data from '%s'...\n", filePath)

		// 1. Read YAML configuration
		yamlFile, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}

		var dataConfig DataConfig
		err = yaml.Unmarshal(yamlFile, &dataConfig)
		if err != nil {
			log.Fatalf("Failed to parse YAML: %v", err)
		}

		// 2. Connect to MongoDB
		observer := kosmos.SummonObserverFor(observation.PurposeAffinityAdmin)
		defer observer.Close()

		// 3. Process the configuration
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		for _, dbConfig := range dataConfig.Databases {
			fmt.Printf("Processing database: %s\n", dbConfig.Name)
			db := observer.Database(dbConfig.Name)

			for _, collConfig := range dbConfig.Collections {
				fmt.Printf("  Processing collection: %s\n", collConfig.Name)

				// Optional: explicitly create collection if it doesn't exist
				fmt.Printf("    -> Calling db.CreateCollection for %s...\n", collConfig.Name)
				err := db.CreateCollection(ctx, collConfig.Name)
				fmt.Printf("    <- db.CreateCollection returned: %v\n", err)

				collection := db.Collection(collConfig.Name)

				// Create indexes
				for _, indexConfig := range collConfig.Indexes {
					if len(indexConfig.Keys) == 0 {
						continue
					}

					// Build index keys preserving order
					keysDoc := bson.D{}
					for _, k := range indexConfig.Keys {
						keysDoc = append(keysDoc, bson.E{Key: k.Field, Value: k.Order})
					}

					indexModel := mongo.IndexModel{
						Keys: keysDoc,
					}

					if indexConfig.Unique {
						indexModel.Options = options.Index().SetUnique(true)
					}

					fmt.Printf("    -> Calling CreateOne for index %v...\n", keysDoc)
					name, err := collection.Indexes().CreateOne(ctx, indexModel)
					fmt.Printf("    <- CreateOne returned name: %s, error: %v\n", name, err)
					if err != nil {
						log.Fatalf("    Failed to create index on collection %s: %v", collConfig.Name, err)
					}
					fmt.Printf("    Created index: %s\n", name)
				}
			}
		}

		for _, userConfig := range dataConfig.Users {
			fmt.Printf("Processing user: %s\n", userConfig.Name)
			password, err := kosmos.CollapseSecret(userConfig.SecretSrc)
			if err != nil {
				log.Fatalf("Failed to extract password source : %v", err)
			}

			responsibility := observation.MemberResponsibility{
				ReadDbs:      userConfig.Read,
				ReadWriteDbs: userConfig.ReadWrite,
				CreatorDbs:   userConfig.Creator,
				IsAdmin:      userConfig.Admin,
			}

			err = observer.UpdateMember(userConfig.Name, password, responsibility, true)
			if err != nil {
				log.Fatalf("Failed to set user: %v", err)
			}
		}

		fmt.Println("Data initialization completed successfully.")
	},
}

func init() {
	dbCmd.AddCommand(dbDeclareCmd)

	// Define persistent flags or local flags
	dbDeclareCmd.Flags().StringP("file", "f", "", "Path to the data initialization YAML file")
	dbDeclareCmd.MarkFlagRequired("file")
}
