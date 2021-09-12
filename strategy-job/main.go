package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vikjdk7/Algotrading-Golang-Rest/strategy-job/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func calculateTodaysStrategyProfits() error {

	t := time.Now().UTC()
	t1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, t.Nanosecond(), t.Location())
	todaysDate := t1.Format(time.RFC3339)
	fmt.Println("Getting Deals closed on todays Date: ", todaysDate)

	cur, err := dealsdb.Find(context.TODO(), bson.M{"closed_at": bson.M{"$gte": todaysDate}})
	if err != nil {
		return errors.New(err.Error())
	}

	// Close the cursor once finished
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		var deal models.Deal
		var strategy_Profits models.Strategy_Profits

		err := cur.Decode(&deal) // decode similar to deserialize process.
		if err != nil {
			return errors.New(err.Error())
		}
		fmt.Println("Processing Deal: ", deal)

		err = strategy_profitdb.FindOne(context.TODO(), bson.M{"strategy_id": deal.StrategyId}).Decode(&strategy_Profits)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return errors.New(err.Error())
			}
		}

		profitValue := strategy_Profits.ProfitValue + deal.ProfitValue
		fmt.Println("Profit Value: ", profitValue)
		opts := options.Update().SetUpsert(true)
		filter := bson.M{}
		filter["strategy_id"] = deal.StrategyId
		update := bson.M{"strategy_id": deal.StrategyId, "user_id": deal.UserId, "created_at": todaysDate, "profit_value": profitValue}

		_, err = strategy_profitdb.UpdateOne(context.TODO(), filter, bson.M{"$set": update}, opts)
		if err != nil {
			return errors.New(err.Error())
		}
	}

	//Delete records earlier than 5 days
	fiveDaysBefore := t1.AddDate(0, 0, -5).Format(time.RFC3339)
	_, err = strategy_profitdb.DeleteMany(context.TODO(), bson.M{"created_at": bson.M{"$lt": fiveDaysBefore}})

	return nil
}

var db *mongo.Client
var strategy_profitdb *mongo.Collection
var exchangedb *mongo.Collection
var strategydb *mongo.Collection
var dealsdb *mongo.Collection
var mongoCtx context.Context

func main() {
	//Uncomment to run locally
	//os.Setenv("MONGODB_URL", "mongodb://127.0.0.1:27017")

	MONGODB_URL := os.Getenv("MONGODB_URL")

	// Initialize MongoDb client
	fmt.Println("Connecting to MongoDB...")

	// non-nil empty context
	mongoCtx = context.Background()
	// Connect takes in a context and options, the connection URI is the only option we pass for now
	db, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(MONGODB_URL))
	// Handle potential errors
	if err != nil {
		log.Fatal(err)
	}

	// Check whether the connection was succesful by pinging the MongoDB server
	err = db.Ping(mongoCtx, nil)
	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %v\n", err)
	} else {
		fmt.Println("Connected to Mongodb")
	}
	// Bind our collection to our global variable for use in other methods
	mongoDb := db.Database("hedgina_algobot")
	strategy_profitdb = mongoDb.Collection("strategy_profit")
	strategydb = mongoDb.Collection("strategy")
	exchangedb = mongoDb.Collection("exchange")
	dealsdb = mongoDb.Collection("deal")

	err = calculateTodaysStrategyProfits()
	if err != nil {
		db.Disconnect(mongoCtx)
		log.Fatalf("%v\n", err)
		//fmt.Println(err)
	}

	db.Disconnect(mongoCtx)
	fmt.Println("Disconnected from Mongo")
}
