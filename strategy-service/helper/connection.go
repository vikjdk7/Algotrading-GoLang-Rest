package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() (*mongo.Collection, *mongo.Collection, *mongo.Collection, *mongo.Collection, *mongo.Collection, *mongo.Collection) {

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

	mongoDB := client.Database("hedgina_algobot")

	strategyCollection := mongoDB.Collection("strategy")
	eventHistoryCollection := mongoDB.Collection("eventhistory_strategy")
	strategy_revisionsCollection := mongoDB.Collection("strategy_revisions")
	dealsCollection := mongoDB.Collection("deals")
	exchangeCollection := mongoDB.Collection("exchange")
	assetsCollection := mongoDB.Collection("assets")

	return strategyCollection, eventHistoryCollection, strategy_revisionsCollection, dealsCollection, exchangeCollection, assetsCollection
}

type ErrorResponse struct {
	StatusCode   int    `json:"status"`
	ErrorMessage string `json:"message"`
}

func GetError(err error, w http.ResponseWriter) {

	//log.Fatal(err.Error())
	var response = ErrorResponse{
		ErrorMessage: err.Error(),
		StatusCode:   http.StatusInternalServerError,
	}

	message, _ := json.Marshal(response)

	w.WriteHeader(response.StatusCode)
	w.Write(message)
}
