/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tieubaoca/chatbot-be/config"
	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
	"github.com/tieubaoca/chatbot-be/utils"
)

// uploadDocumentCmd represents the uploadDocument command
var uploadDocumentCmd = &cobra.Command{
	Use:   "upload-document",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig("config/config.yaml")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		filePath, _ := cmd.Flags().GetString("file")
		tags, _ := cmd.Flags().GetStringArray("tags")
		reinit, _ := cmd.Flags().GetBool("reinit")
		if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
			log.Fatalf("Failed to create upload directory: %v", err)
		}
		// Copy file to upload directory with timestamp
		destPath, err := utils.CopyFileWithTimestamp(filePath, cfg.UploadDir)
		if err != nil {
			log.Fatalf("Failed to copy file: %v", err)
		}

		pdfService := service.NewPDFService(service.DefaultDocumentServiceConfig)

		weaviateDb, err := database.NewWeaviateStore(cfg.WeaviateStoreConfig)
		if err != nil {
			log.Fatalf("Failed to connect to Weaviate database: %v", err)
		}
		if reinit {
			err := weaviateDb.ReInit()
			if err != nil {
				log.Fatalf("Failed to reinitialize Weaviate database: %v", err)
			}
		}

		chunkChan := make(chan types.DocumentChunk)
		req := types.UploadRequest{
			Title: service.GetFileNameWithoutExt(filePath),
			Tags:  tags,
		}
		go pdfService.ProcessPDF(destPath, req, chunkChan)
		// testFile, err := os.Create("test.txt")
		if err != nil {
			log.Fatalf("Failed to create test file: %v", err)
		}
		for chunk := range chunkChan {
			document := &types.Document{
				Content: chunk.Content,
				Metadata: types.Metadata{
					Title: chunk.Metadata.Title,
					Tags:  tags,
					Custom: map[string]string{
						"page": fmt.Sprintf("%d", chunk.Metadata.PageNum),
					},
				},
				CreatedAt: time.Now().Unix(),
			}
			// test to write data to a text file for test

			// defer testFile.Close()
			// testFile.WriteString(document.Content)
			// end test
			err = weaviateDb.UpsertDocument(context.Background(), document, nil)
			if err != nil {
				log.Fatalf("Failed to upload document to Weaviate database: %v", err)
			}
			fmt.Println("Uploaded document page", chunk.Metadata.PageNum)
		}
	},
}

func init() {
	rootCmd.AddCommand(uploadDocumentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadDocumentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uploadDocumentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	uploadDocumentCmd.Flags().StringP("file", "f", "", "Path to the file to upload")
	uploadDocumentCmd.Flags().StringP("database-url", "d", "http://192.168.1.2:8080", "URL for the Weaviate database")
	uploadDocumentCmd.Flags().StringP("text2vec", "t", "text2vec-transformers", "Text2Vec model to use for the AI service")
	uploadDocumentCmd.Flags().BoolP("reinit", "r", false, "Reinitialize the database")
	uploadDocumentCmd.Flags().StringArrayP("tags", "g", []string{}, "Tags for the document")
	uploadDocumentCmd.Flags().StringP("upload-dir", "u", "upload", "Directory to store uploaded files")
	uploadDocumentCmd.Flags().StringP("embed-model", "e", "mxbai-embed-large", "Embedding model to use for the AI service")
}
