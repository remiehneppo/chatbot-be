/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	services "github.com/tieubaoca/chatbot-be/service"
)

// startServerCmd represents the startServer command
var startServerCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the WebSocket chat server",
	Long:  `Starts a WebSocket server that handles AI chat connections`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		baseURL, _ := cmd.Flags().GetString("base-url")
		apiKey := os.Getenv("OPENAI_API_KEY")

		// Initialize services
		aiService := services.NewOpenAIService(baseURL, apiKey)
		wsService := services.NewWebSocketService(aiService)

		// Setup routes
		http.HandleFunc("/ws/chat", wsService.HandleChat)
		http.HandleFunc("/health", wsService.Health)

		fmt.Printf("Starting WebSocket server on port %s...\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startServerCmd)
	startServerCmd.Flags().StringP("port", "p", "8888", "Port to run the server on")
	startServerCmd.Flags().StringP("base-url", "u", "http://localhost:1234/v1", "Base URL for the AI service")
}
