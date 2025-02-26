package config

type Config struct {
	GeminiKeys []string `json:"gemini_keys"`
	Model      string   `json:"model"`
}

var AppConfig Config
