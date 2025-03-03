/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
)

// batchUploadDocumentCmd represents the batchUploadDocument command
var batchUploadDocumentCmd = &cobra.Command{
	Use:   "batch-upload-document",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		databaseURL, _ := cmd.Flags().GetString("database-url")
		text2vec := cmd.Flag("text2vec").Value.String()
		directory, _ := cmd.Flags().GetString("directory")
		//get array tags
		tags, _ := cmd.Flags().GetStringArray("tags")
		reinit, _ := cmd.Flags().GetBool("reinit")

		pdfService := service.NewPDFService(
			types.DocumentServiceConfig{
				MaxChunkSize: 5000,
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
		log.Println("api key", os.Getenv("WEAVIATE_APIKEY"))
		weaviateDb, err := database.NewWeaviateStore(httpScheme, databaseURL, os.Getenv("WEAVIATE_APIKEY"), text2vec)
		if err != nil {
			log.Fatalf("Failed to connect to Weaviate database: %v", err)
		}
		if reinit {
			err := weaviateDb.ReInit()
			if err != nil {
				log.Fatalf("Failed to reinitialize Weaviate database: %v", err)
			}
		}

		// read all pdf files in the directory
		files, err := os.ReadDir(directory)
		if err != nil {
			log.Fatalf("Failed to read directory: %v", err)
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			filePath := fmt.Sprintf("%s/%s", directory, file.Name())
			err := upload(filePath, weaviateDb, pdfService, tags)
			if err != nil {
				log.Println("Failed to upload document %s: %v", filePath, err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(batchUploadDocumentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchUploadDocumentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchUploadDocumentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	batchUploadDocumentCmd.Flags().String("directory", "", "Path to the dir to upload")
	batchUploadDocumentCmd.Flags().StringP("database-url", "d", "http://192.168.1.2:8080", "URL for the Weaviate database")
	batchUploadDocumentCmd.Flags().StringP("text2vec", "t", "text2vec-transformers", "Text2Vec model to use for the AI service")
	batchUploadDocumentCmd.Flags().BoolP("reinit", "r", false, "Reinitialize the database")
	batchUploadDocumentCmd.Flags().StringArrayP("tags", "g", []string{}, "Tags for the document")
}

func upload(filePath string, weaviateDb *database.WeaviateStore, pdfService *service.PDFService, tags []string) error {
	chunkChan := make(chan types.DocumentChunk)
	defer close(chunkChan)
	go pdfService.ProcessPDF(filePath, chunkChan)
	for chunk := range chunkChan {
		document := &database.Document{
			Content: chunk.Content,
			Metadata: database.Metadata{
				Title: chunk.Metadata.Title,
				Tags:  tags,
				Custom: map[string]string{
					"page": fmt.Sprintf("%d", chunk.Metadata.PageNum),
				},
			},
		}
		err := weaviateDb.UpsertDocument(context.Background(), document, nil)
		if err != nil {
			log.Println("Failed to upload document to Weaviate database: %v", err)
			return err
		}
		fmt.Println("Uploaded document page", chunk.Metadata.PageNum)
	}
	return nil
}
