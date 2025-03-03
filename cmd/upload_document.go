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
		databaseURL, _ := cmd.Flags().GetString("database-url")
		text2vec := cmd.Flag("text2vec").Value.String()
		filePath, _ := cmd.Flags().GetString("file")
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

		chunkChan := make(chan types.DocumentChunk)
		req := types.UploadRequest{
			Title: service.GetFileNameWithoutExt(filePath),
			Tags:  tags,
		}
		go pdfService.ProcessPDF(filePath, req, chunkChan)
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

}
