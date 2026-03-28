package shared

import (
	"context"
	"fmt"
	"os"

	dotenvsecret "github.com/borghives/dotenvsecret/go-dotenvsecret"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SiteConfig struct {
	ProjectID         string `mapstructure:"PROJECT_ID"`
	ProxyAddress      string `mapstructure:"ALL_PROXY"`
	MongoDBCmdUri     string
	MongoDBUri        string `mapstructure:"MONGODB_URI"`
	MongoDBAdminUri   string `mapstructure:"MONGODB_ADMIN_URI"`
	MongoDBCreatorUri string `mapstructure:"MONGODB_CREATOR_URI"`
}

func (c *SiteConfig) MergeFromCmd(cmd *cobra.Command) *SiteConfig {
	if cmd == nil {
		return c
	}

	if flag := cmd.Flags().Lookup("uri"); flag != nil {
		viper.BindPFlag("MONGODB_URI", flag)
		c.MongoDBCmdUri = flag.Value.String()
	}
	if flag := cmd.Flags().Lookup("project"); flag != nil {
		viper.BindPFlag("PROJECT_ID", flag)
	}

	viper.Unmarshal(c)

	return c
}

func (c *SiteConfig) MergeFromFile(name string) *SiteConfig {
	viper.SetConfigFile(name)
	_ = viper.MergeInConfig()
	viper.Unmarshal(c)
	return c
}

func (c *SiteConfig) FindPrivilegeMongoUri(creator bool) string {
	if c.MongoDBCmdUri != "" {
		return c.MongoDBCmdUri
	}

	if creator && c.MongoDBCreatorUri != "" {
		return c.MongoDBCreatorUri
	}

	if c.MongoDBAdminUri != "" {
		return c.MongoDBAdminUri
	}

	if c.MongoDBCreatorUri != "" {
		return c.MongoDBCreatorUri
	}

	return c.MongoDBUri
}

func LoadSiteConfig() (config SiteConfig, err error) {
	// 1. Tell Viper where to look for the file
	viper.AddConfigPath(".")    // Search in the current working directory
	viper.SetConfigName(".env") // Look for a file named ".env"
	viper.SetConfigType("env")  // Treat the file as a .env format

	// 3. Read the file
	if err = viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist in production
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	projectId := os.Getenv("PROJECT_ID")
	if projectId == "" {
		projectId = viper.GetString("PROJECT_ID")
		if projectId != "" {
			os.Setenv("PROJECT_ID", projectId)
		}
	}

	if projectId != "" {
		dotenvsecret.Load(context.Background(), dotenvsecret.NewGCPSecretManager())
	}

	// 2. Enable environment variable overrides
	// This is crucial for Docker/Kubernetes production environments
	viper.BindEnv("ALL_PROXY")
	viper.BindEnv("PROJECT_ID")
	viper.BindEnv("MONGODB_URI")
	viper.BindEnv("MONGODB_ADMIN_URI")
	viper.BindEnv("MONGODB_CREATOR_URI")
	viper.AutomaticEnv()

	// 4. Unmarshal the values into our struct
	err = viper.Unmarshal(&config)
	return
}

func GetProjectParents() string {
	projectID := viper.GetString("PROJECT_ID")

	if projectID == "" {
		fmt.Println("Project flag and environment PROJECT_ID is not set")
		return ""
	}
	return fmt.Sprintf("projects/%s", projectID)
}
