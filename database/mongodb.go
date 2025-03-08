package database

import (
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var DefaultMongoClient *mongo.Client

func init() {
	mongoClient, err := mongo.Connect(options.Client().
		ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Println("Failed to connect to MongoDB:", err)
	}
	DefaultMongoClient = mongoClient
}
