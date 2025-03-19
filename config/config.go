package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var OllamaT2VModuleConfig = map[string]interface{}{
	"text2vec-ollama": map[string]interface{}{
		"apiEndpoint": "http://host.docker.internal:11434",
		"model":       "hf.co/sheldonrobinson/all-MiniLM-L12-v2-Q4_K_M-GGUF",
	},
}
var OllamaGenerativeConfig = map[string]interface{}{
	"generative-ollama": map[string]interface{}{
		"apiEndpoint": "http://host.docker.internal:11434",
	},
}

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
	OllamaGenerativeConfig["generative-ollama"].(map[string]interface{})["model"] = config.Model
	config.WeaviateStoreConfig.ModuleConfig = OllamaGenerativeConfig

	if config.WeaviateStoreConfig.Text2Vec == "text2vec-ollama" {
		for k, v := range OllamaT2VModuleConfig["text2vec-ollama"].(map[string]interface{}) {
			config.WeaviateStoreConfig.ModuleConfig[k] = v
		}
	}

	return &config, nil
}
