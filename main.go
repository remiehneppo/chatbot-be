package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai/jsonschema"
	services "github.com/tieubaoca/chatbot-be/service"
)

func main() {
	openaiService := services.NewOpenAIService("http://localhost:1234/v1", "")
	openaiService.RegisterFunctionCall(
		"search",
		"search on the internet",
		jsonschema.Definition{
			Type: "object",
			Properties: map[string]jsonschema.Definition{
				"query": {
					Type:        jsonschema.String,
					Description: "The search query",
				},
			},
			Required: []string{"query"},
		},
		func(ctx context.Context, args []byte) (any, error) {
			return "This is an AI company", nil
		},
	)
	messages := []services.Message{
		{
			Role:    "user",
			Content: "Search on the internet about deepseek company then search about chatgpt then summarize it",
		},
	}
	err := openaiService.ChatStream(context.Background(), messages,
		func(response string) {
			fmt.Print(response)
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
