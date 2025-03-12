/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/tieubaoca/chatbot-be/config"
	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/handler"
	"github.com/tieubaoca/chatbot-be/middleware"
	"github.com/tieubaoca/chatbot-be/repository"
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

		mongoClient := database.DefaultMongoClient

		if err := mongoClient.Ping(context.Background(), nil); err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		mongoDb := mongoClient.Database("chatbot")

		//init repo
		userRepo := repository.NewUserRepo(mongoDb.Collection("users"))
		//init service
		userService := service.NewUserService(userRepo)
		uploadService := service.NewFileService(cfg.UploadDir, weaviateDb, pdfService)

		// Initialize handlers
		corsHandler := handler.NewCorsHandler()
		uploadHandler := handler.NewUploadHandler(uploadService)
		chatHandler := handler.NewChatHandler(aiService)
		searchHandler := handler.NewSearchHandler(weaviateDb)
		pdfHandler := handler.NewDocumentHandler(cfg.UploadDir) // Add this line
		loginHandler := handler.NewLoginHandler(userService)

		userMngHandler := handler.NewUserManageHandler(userService)
		// Setup routes
		// user request
		userMux := http.NewServeMux()
		userMux.Handle("/chat", chatHandler.HandleChat())
		userMux.Handle("/documents/search", searchHandler.HandleSearch())
		userMux.Handle("/documents/ask-ai", searchHandler.HandleAskAI())
		userMux.Handle("/pdf", pdfHandler.ServeDocument()) // Add this line

		// admin request
		adminMux := http.NewServeMux()
		adminMux.Handle("/upload", uploadHandler.UploadDocumentHandler())
		adminMux.Handle("/users/create", userMngHandler.HandleCreateUser())
		adminMux.Handle("/users/batch-create", userMngHandler.HandlerBatchCreateUser())
		adminMux.Handle("/users/paginate", userMngHandler.HandlePaginateUser())
		adminMux.Handle("/users/get", userMngHandler.HandleGetUser())
		adminMux.Handle("/users/update", userMngHandler.HandleUpdateUser())
		adminMux.Handle("/users/delete", userMngHandler.HandleDeleteUser())

		mux := http.NewServeMux()
		mux.Handle("/api/v1/", middleware.AuthMiddleware(userMux))
		mux.Handle("/api/v1/login", loginHandler.HandleLogin())
		mux.Handle("/admin/api/v1/", middleware.AdminAuthMiddleware(adminMux))

		http.Handle("/", corsHandler.CorsMiddleware(mux))

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
