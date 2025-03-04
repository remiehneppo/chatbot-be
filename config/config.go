package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port                string              `mapstructure:"port"`
	AIEndpoint          string              `mapstructure:"ai_endpoint"`
	Model               string              `mapstructure:"model"`
	OpenAIAPIKey        string              `mapstructure:"OPENAI_API_KEY"`
	UploadDir           string              `mapstructure:"upload_dir"`
	WeaviateStoreConfig WeaviateStoreConfig `mapstructure:"weaviate_store_config"`
}

type WeaviateStoreConfig struct {
	Host         string       `mapstructure:"host"`
	APIKey       string       `mapstructure:"WEAVIATE_APIKEY"` // Changed to match env var
	Text2Vec     string       `mapstructure:"text2vec"`
	ModuleConfig ModuleConfig `mapstructure:"module_config"`
}

type ModuleConfig map[string]interface{}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set up Viper to read from config file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Set up Viper to read from environment variables
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Bind environment variables
	v.BindEnv("OPENAI_API_KEY")
	v.BindEnv("WEAVIATE_APIKEY")

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
