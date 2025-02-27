/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/handler"
	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
)

// startServerCmd represents the startServer command
var startServerCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the chat server",
	Long:  `Starts a server that handles AI chat connections`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		baseURL, _ := cmd.Flags().GetString("base-url")
		model := cmd.Flag("model").Value.String()
		databaseURL, _ := cmd.Flags().GetString("database-url")
		text2vec := cmd.Flag("text2vec").Value.String()
		apiKey := os.Getenv("OPENAI_API_KEY")

		// Initialize services

		pdfService := service.NewPDFService(
			types.DocumentServiceConfig{
				MaxChunkSize: 1000,
				OverlapSize:  100,
			})
		var httpScheme string
		if strings.Contains(databaseURL, "https") {
			httpScheme = "https"
			databaseURL = strings.Replace(databaseURL, "https://", "", 1)
		} else {
			httpScheme = "http"
			databaseURL = strings.Replace(databaseURL, "http://", "", 1)
		}
		weaviateDb, err := database.NewWeaviateStore(httpScheme, databaseURL, os.Getenv("WEAVIATE_APIKEY"), text2vec)
		if err != nil {
			log.Fatalf("Failed to connect to Weaviate database: %v", err)
		}
		aiService := service.NewOpenAIService(baseURL, apiKey, model, weaviateDb)
		if err := aiService.RegisterRAGFunctionCall(); err != nil {
			log.Fatalf("Failed to register RAG function call: %v", err)
		}

		// Initialize handlers
		corsHandler := handler.NewCorsHandler()
		uploadService := service.NewFileService("upload", weaviateDb, pdfService)
		uploadHandler := handler.NewUploadHandler(uploadService)
		chatHandler := handler.NewChatHandler(aiService)
		searchHandler := handler.NewSearchHandler(weaviateDb) // Add this line

		// Setup routes
		http.Handle("/api/v1/upload", corsHandler.CorsMiddleware(uploadHandler.UploadDocumentHandler()))
		http.Handle("/api/v1/chat", corsHandler.CorsMiddleware(chatHandler.HandleChat()))
		http.Handle("/api/v1/search", corsHandler.CorsMiddleware(searchHandler.HandleSearch())) // Add this line

		log.Printf("Starting WebSocket server on port %s...\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startServerCmd)
	startServerCmd.Flags().StringP("port", "p", "8888", "Port to run the server on")
	startServerCmd.Flags().StringP("base-url", "u", "http://localhost:1234/v1", "Base URL for the AI service")
	startServerCmd.Flags().String("model", "", "Model to use for the AI service")
	startServerCmd.Flags().StringP("database-url", "d", "http://192.168.1.2:8080", "URL for the Weaviate database")
	startServerCmd.Flags().StringP("text2vec", "t", "text2vec-transformers", "Text2Vec model to use for the AI service")
}
