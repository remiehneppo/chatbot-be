/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
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
		// Setup Gin router
		router := gin.Default()

		// Apply global middleware
		router.Use(corsHandler.CorsMiddleware)

		// API v1 routes - require authentication
		apiV1 := router.Group("/api/v1")
		apiV1.POST("/login", loginHandler.HandleLogin)

		// Protected user routes
		userRoutes := apiV1.Group("/")
		userRoutes.Use(middleware.AuthMiddleware)
		{
			userRoutes.POST("/chat", chatHandler.HandleChat)
			userRoutes.GET("/documents/search", searchHandler.HandleSearch)
			userRoutes.POST("/documents/ask-ai", searchHandler.HandleAskAI)
			userRoutes.GET("/pdf", pdfHandler.ServeDocument)
		}

		// Admin routes - require admin authentication
		adminRoutes := router.Group("/admin/api/v1")
		adminRoutes.Use(middleware.AdminAuthMiddleware)
		{
			adminRoutes.POST("/upload", uploadHandler.UploadDocumentHandler)
			adminRoutes.POST("/users/create", userMngHandler.HandleCreateUser)
			adminRoutes.POST("/users/batch-create", userMngHandler.HandlerBatchCreateUser)
			adminRoutes.GET("/users/paginate", userMngHandler.HandlePaginateUser)
			adminRoutes.GET("/users/get", userMngHandler.HandleGetUser)
			adminRoutes.PUT("/users/update", userMngHandler.HandleUpdateUser)
			adminRoutes.DELETE("/users/delete", userMngHandler.HandleDeleteUser)
		}

		log.Printf("Starting server on port %s...\n", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal("Server error:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startServerCmd)
	startServerCmd.Flags().StringP("config", "c", "config/config.yaml", "config file")
}
