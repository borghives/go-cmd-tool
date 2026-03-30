package cmd

import (
	"fmt"
	"log"

	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/observation"
	"github.com/spf13/cobra"
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
		observer := kosmos.SummonObserverFor(observation.PurposeAffinityAdmin)
		defer observer.Close()

		status, err := observer.Status()
		if err != nil {
			log.Fatalf("Failed to get replica set status: %v", err)
		}

		printSyncStatus(status.RSStatus)
		fmt.Printf("\n")
		printServerHealth(status.ServerStatus)
	},
}

var reVoteCmd = &cobra.Command{
	Use:   "revote",
	Short: "Force MongoDB replica set to revote",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Action: Force MongoDB replica set to revote...\n")
		observer := kosmos.SummonObserverFor(observation.PurposeAffinityAdmin)
		defer observer.Close()

		err := observer.ReelectPrimary()
		if err != nil {
			log.Fatalf("Failed to force election: %v", err)
		}
	},
}

func printSyncStatus(status observation.RSStatus) {
	fmt.Printf("REPLICA SET: %s\n", status.Set)
	fmt.Printf("%-25s %-12s %-25s %s\n", "NAME", "STATE", "OPTIME (DATE)", "SYNC SOURCE")
	fmt.Println("-----------------------------------------------------------------------------------------")

	for _, m := range status.Members {
		syncSource := m.SyncSourceHost
		if syncSource == "" {
			if m.StateStr == "PRIMARY" {
				syncSource = "SELF (Primary)"
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
func printServerHealth(stats observation.ServerStatus) {

	fmt.Printf("--- Server Health ---\n")
	fmt.Printf("Uptime:      %d seconds\n", stats.Uptime)
	fmt.Printf("Connections: %d used / %d available\n", stats.Connections.Current, stats.Connections.Available)
	fmt.Printf("Memory:      %d MB Resident\n", stats.Mem.Resident)
	fmt.Printf("Throughput:  Q:%d I:%d U:%d\n", stats.Opcounters.Query, stats.Opcounters.Insert, stats.Opcounters.Update)
}

func init() {
	rsCmd.AddCommand(rsStatusCmd)
	rsCmd.AddCommand(reVoteCmd)
}
