package shared

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func GetProjectParents(cmd *cobra.Command) string {
	projectID := ""
	if cmd != nil {
		projectID, _ = cmd.Flags().GetString("project")
	}

	if projectID == "" {
		projectID = os.Getenv("PROJECT_ID")
		if projectID == "" {
			fmt.Println("Project flag and environment PROJECT_ID is not set")
			return ""
		}
	}
	return fmt.Sprintf("projects/%s", projectID)
}
