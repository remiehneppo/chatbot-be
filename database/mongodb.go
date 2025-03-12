package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var DefaultMongoClient *mongo.Client

func init() {
	godotenv.Load()
	mongoClient, err := mongo.Connect(options.Client().
		ApplyURI(os.Getenv("MONGODB_URI")).
		SetBSONOptions(
			&options.BSONOptions{
				ObjectIDAsHexString: true,
			},
		))
	if err != nil {
		log.Println("Failed to connect to MongoDB:", err)
	}
	DefaultMongoClient = mongoClient
}
