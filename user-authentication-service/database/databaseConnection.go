package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//DBinstance func
func DBinstance() *mongo.Client {

	//Uncomment to run locally
	os.Setenv("MONGODB_URL", "mongodb://127.0.0.1:27017")

	MONGODB_URL := os.Getenv("MONGODB_URL")

	// Set client options
	clientOptions := options.Client().ApplyURI(MONGODB_URL)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check whether the connection was succesful by pinging the MongoDB server
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %v\n", err)
	} else {
		fmt.Println("Connected to Mongodb")
	}

	return client
}

//Client Database instance
var Client *mongo.Client = DBinstance()

//OpenCollection is a  function makes a connection with a collection in the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database("hedgina_algobot").Collection(collectionName)

	return collection
}
