package configs

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/spf13/viper"
)

type Config struct {
	LogLevel        string        `mapstructure:"LOG_LEVEL"`
	HTTPServerHost  string        `mapstructure:"HTTP_SERVER_HOST"`
	KafkaGroupID    string        `mapstructure:"KAFKA_GROUP_ID"`
	BrokerHost      string        `mapstructure:"BROKER_HOST"`
	MetastoreHost   string        `mapstructure:"METASTORE_HOST"`
	MetastorePort   int           `mapstructure:"METASTORE_PORT"`
	SparkHost       string        `mapstructure:"SPARK_HOST"`
	SparkPort       int           `mapstructure:"SPARK_PORT"`
	RedisURI        string        `mapstructure:"REDIS_URI"`
	DataspaceHost   string        `mapstructure:"DATASPACE_HOST"`
	DataserviceHost string        `mapstructure:"DATASERVICE_HOST"`
	AirbyteConfig   AirbyteConfig `mapstructure:"AIRBYTE_CONFIG"`
}

type AirbyteConfig struct {
	Host                 string               `mapstructure:"HOST"`
	OrganizationId       string               `mapstructure:"ORGANIZATION_ID"`
	SyncWebhook          string               `mapstructure:"SYNC_WEBHOOK"`
	DestinationTemplates DestinationTemplates `mapstructure:"destination_templates"`
}

type DestinationTemplates struct {
	S3 S3 `mapstructure:"S3"`
}

type S3 struct {
	DefinitionId   string                 `mapstructure:"DEFINITION_ID"`
	DockerImageTag string                 `mapstructure:"DOCKER_IMAGE_TAG"`
	BucketKey      string                 `mapstructure:"BUCKET_KEY"`
	Config         map[string]interface{} `mapstructure:"CONFIG"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config.yaml")

	viper.AddConfigPath("configs")

	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("HTTP_SERVER_HOST", ":8080")
	viper.SetDefault("KAFKA_GROUP_ID", "manager")
	viper.SetDefault("AIRBYTE_CONFIG_HOST", "http://localhost:8001/api")
	viper.SetDefault("AIRBYTE_CONFIG_SYNC_WEBHOOK", "http://host.docker.internal:8080/ingestions/airbyte_internal/sync_webhook")
	viper.SetDefault("BROKER_HOST", "localhost:29092")
	viper.SetDefault("METASTORE_HOST", "localhost")
	viper.SetDefault("METASTORE_PORT", 9083)
	viper.SetDefault("SPARK_HOST", "localhost")
	viper.SetDefault("SPARK_PORT", 10000)
	viper.SetDefault("REDIS_URI", "redis://localhost:6379/0")
	viper.SetDefault("DATASPACE_HOST", "http://localhost:3000/pub")
	viper.SetDefault("DATASERVICE_HOST", "http://localhost:9099/api")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return nil, err
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config: %v", err)
	}

	return &config, nil
}

func (c *Config) GetLogger() log.Logger {
	logLevel := strings.ToLower(c.LogLevel)
	parsedLogLevel := level.Allow(level.ParseDefault(logLevel, level.InfoValue()))
	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = level.NewFilter(logger, parsedLogLevel)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}
	return logger
}
