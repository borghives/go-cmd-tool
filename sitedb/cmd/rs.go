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

		fmt.Printf("\n")

		stats, err := getServerHealth(client)
		if err != nil {
			log.Fatalf("Failed to get server health: %v", err)
		}
		printServerHealth(stats)
	},
}

var reVoteCmd = &cobra.Command{
	Use:   "revote",
	Short: "Force MongoDB replica set to revote",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Force MongoDB replica set to revote...\n")
		client := shared.MustConnectAdminDbClient(&config, false)
		defer client.Disconnect(context.Background())

		err := forceElection(client)
		if err != nil {
			log.Fatalf("Failed to force election: %v", err)
		}
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

type ServerStatus struct {
	Uptime      int64 `bson:"uptime"`
	Connections struct {
		Current   int `bson:"current"`
		Available int `bson:"available"`
	} `bson:"connections"`
	Opcounters struct {
		Insert  int64 `bson:"insert"`
		Query   int64 `bson:"query"`
		Update  int64 `bson:"update"`
		Delete  int64 `bson:"delete"`
		Command int64 `bson:"command"`
	} `bson:"opcounters"`
	Mem struct {
		Resident int64 `bson:"resident"` // in MB
		Virtual  int64 `bson:"virtual"`
	} `bson:"mem"`
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
func printServerHealth(stats *ServerStatus) {

	fmt.Printf("--- Server Health ---\n")
	fmt.Printf("Uptime:      %d seconds\n", stats.Uptime)
	fmt.Printf("Connections: %d used / %d available\n", stats.Connections.Current, stats.Connections.Available)
	fmt.Printf("Memory:      %d MB Resident\n", stats.Mem.Resident)
	fmt.Printf("Throughput:  Q:%d I:%d U:%d\n", stats.Opcounters.Query, stats.Opcounters.Insert, stats.Opcounters.Update)
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

func forceElection(client *mongo.Client) error {
	adminDB := client.Database("admin")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Command: replSetStepDown
	// 60 is the 'stepDownSecs' - how long the node stays ineligible to be Primary
	command := bson.D{{Key: "replSetStepDown", Value: 60}}

	err := adminDB.RunCommand(ctx, command).Err()
	if err != nil {
		// Note: The driver often returns an error here because the
		// connection is dropped when the node steps down.
		// This is usually expected behavior.
		return err
	}
	return nil
}

func getServerHealth(client *mongo.Client) (*ServerStatus, error) {
	var stats ServerStatus
	err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "serverStatus", Value: 1}}).Decode(&stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func init() {
	rsCmd.AddCommand(rsStatusCmd)
	rsCmd.AddCommand(reVoteCmd)
}
