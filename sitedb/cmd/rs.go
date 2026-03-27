package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/borghives/go-cmd-tool/shared"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var rsCmd = &cobra.Command{
	Use:   "rs",
	Short: "Manage MongoDB replica set",
}

var rsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get MongoDB replica set status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Get MongoDB replica set status...\n")
		client := shared.MustConnectAdminDbClient(&config, false)
		defer client.Disconnect(context.Background())

		status, err := getReplicaSetStatus(client)
		if err != nil {
			log.Fatalf("Failed to get replica set status: %v", err)
		}

		printSyncStatus(status)
	},
}

// RSStatus defines the top-level response from replSetGetStatus
type RSStatus struct {
	Set     string   `bson:"set"`
	Members []Member `bson:"members"`
}

// Member defines the individual node status in the replica set
type Member struct {
	Name           string    `bson:"name"`
	StateStr       string    `bson:"stateStr"`
	OptimeDate     time.Time `bson:"optimeDate"`
	SyncSourceHost string    `bson:"syncSourceHost"`
	Health         int       `bson:"health"` // 1 for up, 0 for down
}

func printSyncStatus(status *RSStatus) {
	fmt.Printf("REPLICA SET: %s\n", status.Set)
	fmt.Printf("%-25s %-12s %-25s %s\n", "NAME", "STATE", "OPTIME (DATE)", "SYNC SOURCE")
	fmt.Println("-----------------------------------------------------------------------------------------")

	for _, m := range status.Members {
		syncSource := m.SyncSourceHost
		if syncSource == "" {
			if m.StateStr == "PRIMARY" {
				syncSource = "N/A (Primary)"
			} else {
				syncSource = "Unknown/None"
			}
		}

		// Format the optime date for readability
		optimeStr := m.OptimeDate.Format("2006-01-02 15:04:05")

		fmt.Printf("%-25s %-12s %-25s %s\n",
			m.Name,
			m.StateStr,
			optimeStr,
			syncSource,
		)
	}
}

func getReplicaSetStatus(client *mongo.Client) (*RSStatus, error) {
	// rs.status() must be run against the 'admin' database
	adminDB := client.Database("admin")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result RSStatus
	// The command name in the driver is "replSetGetStatus"
	err := adminDB.RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: 1}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func init() {
	rsCmd.AddCommand(rsStatusCmd)
}
