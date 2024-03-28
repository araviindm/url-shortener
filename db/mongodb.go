package db

import (
	"context"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Database

func init() {
	MONGO_URI := os.Getenv("MONGO_URI")

	if MONGO_URI == "" {
		log.Fatal("MONGO_URI environment variable is not set")
	}

	clientOptions := options.Client().ApplyURI(MONGO_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	MongoClient = client.Database("URLShortener")
}
