package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

// TestConfig holds configuration values used in testing
type TestConfig struct {
	OpenAPISchemaPath     string                 `mapstructure:"OPEN_API_SCHEMA_PATH"`
	RestAPIHost           string                 `mapstructure:"REST_API_HOST"`
	AirbyteAPIHost        string                 `mapstructure:"AIRBYTE_API_HOST"`
	MetastoreHost         string                 `mapstructure:"METASTORE_HOST"`
	MetastorePort         int                    `mapstructure:"METASTORE_PORT"`
	DefaultOrganizationID string                 `mapstructure:"DEFAULT_ORGANIZATION_ID"`
	E2ESourceDefID        string                 `mapstructure:"E2E_SOURCE_DEF_ID"`
	S3DestDefID           string                 `mapstructure:"S3_DEST_DEF_ID"`
	S3DestDefImgTag       string                 `mapstructure:"S3_DEST_DEF_IMG_TAG"`
	S3MinioConfig         map[string]interface{} `mapstructure:"S3_MINIO_CONFIG"`
}

// LoadTestConfig initializes Viper and loads the test configuration values
func LoadTestConfig() (*TestConfig, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config.yaml")

	viper.AddConfigPath("./configs")

	viper.AutomaticEnv()

	// Set default values for testing
	viper.SetDefault("OPEN_API_SCHEMA_PATH", "../../docs/open-api-schema.yaml")
	viper.SetDefault("REST_API_HOST", "http://localhost:8080")
	viper.SetDefault("AIRBYTE_API_HOST", "http://localhost:8000/api")
	viper.SetDefault("MINIO_HOST", "http://localhost:9000")
	viper.SetDefault("METASTORE_HOST", "localhost")
	viper.SetDefault("METASTORE_PORT", 9083)
	viper.SetDefault("DEFAULT_ORGANIZATION_ID", "00000000-0000-0000-0000-000000000000")
	viper.SetDefault("E2E_SOURCE_DEF_ID", "d53f9084-fa6b-4a5a-976c-5b8392f4ad8a")
	viper.SetDefault("S3_DEST_DEF_ID", "4816b78f-1489-44c1-9060-4b19d5fa9362")
	viper.SetDefault("S3_DEST_DEF_IMG_TAG", "1.3.0")

	// Read in the test config file, if available
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("No config file found or invalid config file for tests: %v\n", err)
	}

	// Unmarshal configuration into struct
	var testConfig TestConfig
	err := viper.Unmarshal(&testConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to decode test config: %v", err)
	}

	return &testConfig, nil
}
