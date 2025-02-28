/*
Copyright © 2025 tieubaoca
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tieubaoca/chatbot-be/database"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chatbot-be",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		baseURL, _ := cmd.Flags().GetString("base-url")
		model := cmd.Flag("model").Value.String()
		_ = model
		databaseURL, _ := cmd.Flags().GetString("database-url")
		text2vec := cmd.Flag("text2vec").Value.String()
		apiKey := os.Getenv("OPENAI_API_KEY")
		dbApiKey := os.Getenv("WEAVIATE_APIKEY")

		var httpScheme string
		if strings.Contains(databaseURL, "https") {
			httpScheme = "https"
			databaseURL = strings.Replace(databaseURL, "https://", "", 1)
		} else {
			httpScheme = "http"
			databaseURL = strings.Replace(databaseURL, "http://", "", 1)
		}
		weaviateConfig := weaviate.Config{
			Host:   databaseURL,
			Scheme: httpScheme,
			AuthConfig: auth.ApiKey{
				Value: dbApiKey,
			},
			Headers: map[string]string{
				"X-Weaviate-Api-Key":     apiKey,
				"X-Weaviate-Cluster-Url": fmt.Sprintf("%s://%s", httpScheme, databaseURL),
				"X-OpenAI-BaseURL":       baseURL,
				"X-OpenAI-Api-Key":       apiKey,
			},
		}

		weaviateClient, err := weaviate.NewClient(weaviateConfig)
		if err != nil {
			fmt.Println("Failed to create Weaviate client: ", err)
			os.Exit(1)
		}
		// delete class Document
		err = weaviateClient.Schema().ClassDeleter().WithClassName("Document").Do(context.Background())
		if err != nil {
			fmt.Println("Failed to delete class Document: ", err)
			os.Exit(1)
		}
		classObj := database.DOCUMENT_CLASS_OBJECT
		classObj.Vectorizer = text2vec
		classObj.ModuleConfig = map[string]interface{}{
			"qna-openai": map[string]interface{}{
				"model": model,
			},
		}
		err = weaviateClient.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
		if err != nil {
			fmt.Println("Failed to create class Document: ", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chatbot-be.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	rootCmd.Flags().StringP("base-url", "u", "http://localhost:1234/v1", "Base URL for the AI service")
	rootCmd.Flags().String("model", "", "Model to use for the AI service")
	rootCmd.Flags().StringP("database-url", "d", "http://192.168.1.2:8080", "URL for the Weaviate database")
	rootCmd.Flags().StringP("text2vec", "t", "text2vec-transformers", "Text2Vec model to use for the AI service")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".chatbot-be" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".chatbot-be")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
