package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
)

var MONGO *mongo.Client

func ConnectMongo() {
	println("Initializing MONGODB connection")
	uri := os.Getenv("MONGODB_URL")

	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable.")
	}
	MONGO, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	database := os.Getenv("MONGODB_DATABASE_INITIAL")
	if database == "" {
		log.Fatal("Set your 'MONGODB_DATABASE_INITIAL' environment variable.")
	}
	err = MONGO.Database(database).CreateCollection(context.Background(), "messages")

	if err == nil {
		fmt.Printf("Successfully opened mongodb connection at %s, connected to the database %s(%s)", uri, database, "messages\n")
	}
}

func Close() {
	MONGO.Disconnect(context.TODO())
}
