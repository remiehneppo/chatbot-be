/*
Copyright © 2025 tieubaoca
*/
package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/tieubaoca/chatbot-be/cmd"
)

func main() {
	cmd.Execute()
}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
