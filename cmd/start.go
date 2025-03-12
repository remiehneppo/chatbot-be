/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/tieubaoca/chatbot-be/config"
	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/handler"
	"github.com/tieubaoca/chatbot-be/middleware"
	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
)

// startServerCmd represents the startServer command
var startServerCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the chat server",
	Long:  `Starts a server that handles AI chat connections`,
	Run: func(cmd *cobra.Command, args []string) {

		cfg, err := config.LoadConfig("config/config.yaml")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		// Initialize services

		pdfService := service.NewPDFService(
			types.DocumentServiceConfig{
				MaxChunkSize: 1000,
				OverlapSize:  100,
			})

		weaviateDb, err := database.NewWeaviateStore(cfg.WeaviateStoreConfig)
		if err != nil {
			log.Fatalf("Failed to connect to Weaviate database: %v", err)
		}
		aiService := service.NewOpenAIService(cfg.AIEndpoint, cfg.OpenAIAPIKey, cfg.Model, weaviateDb)
		if err := aiService.RegisterRAGFunctionCall(); err != nil {
			log.Fatalf("Failed to register RAG function call: %v", err)
		}

		// Initialize handlers
		corsHandler := handler.NewCorsHandler()
		uploadService := service.NewFileService(cfg.UploadDir, weaviateDb, pdfService)
		uploadHandler := handler.NewUploadHandler(uploadService)
		chatHandler := handler.NewChatHandler(aiService)
		searchHandler := handler.NewSearchHandler(weaviateDb)
		pdfHandler := handler.NewDocumentHandler(cfg.UploadDir) // Add this line
		// Setup routes
		// user request
		userMux := http.NewServeMux()
		userMux.Handle("/chat", corsHandler.CorsMiddleware(chatHandler.HandleChat()))
		userMux.Handle("/documents/search", corsHandler.CorsMiddleware(searchHandler.HandleSearch()))
		userMux.Handle("/documents/ask-ai", corsHandler.CorsMiddleware(searchHandler.HandleAskAI()))
		userMux.Handle("/pdf", corsHandler.CorsMiddleware(pdfHandler.ServeDocument())) // Add this line

		// admin request
		adminMux := http.NewServeMux()
		adminMux.Handle("/upload", corsHandler.CorsMiddleware(uploadHandler.UploadDocumentHandler()))

		http.Handle("/api/v1", middleware.AuthMiddleware(userMux))
		http.Handle("/admin/api/v1", middleware.AdminAuthMiddleware(adminMux))

		log.Printf("Starting WebSocket server on port %s...\n", cfg.Port)
		if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startServerCmd)
	startServerCmd.Flags().StringP("config", "c", "config/config.yaml", "config file")
}
