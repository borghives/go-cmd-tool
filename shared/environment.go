package shared

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SiteConfig struct {
	ProjectID      string `mapstructure:"PROJECT_ID"`
	MongoDBUri     string `mapstructure:"MONGODB_URI"`
	MongoDBAuthUri string `mapstructure:"MONGODB_AUTH_URI"`
}

func (c SiteConfig) OverrideFromCmd(cmd *cobra.Command) *SiteConfig {
	if cmd == nil {
		return &c
	}

	mongoDBUriFmt, _ := cmd.Flags().GetString("uri")
	if mongoDBUriFmt != "" {
		c.MongoDBUri = mongoDBUriFmt
	}

	return &c
}

func LoadSiteConfig() (config SiteConfig, err error) {
	// 1. Tell Viper where to look for the file
	viper.AddConfigPath(".")    // Search in the current working directory
	viper.SetConfigName(".env") // Look for a file named ".env"
	viper.SetConfigType("env")  // Treat the file as a .env format

	// 2. Enable environment variable overrides
	// This is crucial for Docker/Kubernetes production environments
	viper.AutomaticEnv()

	// 3. Read the file
	if err = viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist in production
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	// 4. Unmarshal the values into our struct
	err = viper.Unmarshal(&config)
	return
}

func GetProjectParents(cfg *SiteConfig) string {
	projectID := os.Getenv("PROJECT_ID")

	if cfg != nil {
		projectID = cfg.ProjectID
	}

	if projectID == "" {
		fmt.Println("Project flag and environment PROJECT_ID is not set")
		return ""
	}
	return fmt.Sprintf("projects/%s", projectID)
}
