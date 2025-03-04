/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tieubaoca/chatbot-be/config"
	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
	"github.com/tieubaoca/chatbot-be/utils"
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
		reinit, _ := cmd.Flags().GetBool("reinit")
		directory, _ := cmd.Flags().GetString("directory")
		tags, _ := cmd.Flags().GetStringArray("tags")

		cfg, err := config.LoadConfig("config/config.yaml")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		pdfService := service.NewPDFService(
			types.DocumentServiceConfig{
				MaxChunkSize: 5000,
				OverlapSize:  100,
			})

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

		// read all pdf files in the directory
		files, err := os.ReadDir(directory)
		if err != nil {
			log.Fatalf("Failed to read directory: %v", err)
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			destPath, err := utils.CopyFileWithTimestamp(filepath.Join(directory, file.Name()), cfg.UploadDir)
			if err != nil {
				log.Printf("Failed to copy file %s: %v", file, err)
				continue
			}
			err = upload(destPath, weaviateDb, pdfService, tags)
			if err != nil {
				log.Printf("Failed to upload document %s: %v", destPath, err)
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

	batchUploadDocumentCmd.Flags().BoolP("reinit", "r", false, "Reinitialize the database")
	batchUploadDocumentCmd.Flags().StringP("directory", "d", "", "Path to the directory containing PDF files")
	batchUploadDocumentCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags to add to the document")
}

func upload(filePath string, weaviateDb *database.WeaviateStore, pdfService *service.PDFService, tags []string) error {
	chunkChan := make(chan types.DocumentChunk)
	req := types.UploadRequest{
		Title: service.GetFileNameWithoutExt(filePath),
		Tags:  tags,
	}
	defer close(chunkChan)
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
		err := weaviateDb.UpsertDocument(context.Background(), document, nil)
		if err != nil {
			log.Printf("Failed to upload document to Weaviate database: %v", err)
			return err
		}
		fmt.Println("Uploaded document page", chunk.Metadata.PageNum)
	}
	return nil
}
