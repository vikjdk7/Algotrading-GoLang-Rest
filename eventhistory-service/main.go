package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/eventhistory-service/helper"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/eventhistory-service/middleware"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/eventhistory-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getEventHistoryStrategy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var customError models.ErrorString

	token := r.Header.Get("token")

	if token == "" {
		customError.S = "Token cannot be empty"
		helper.GetError(&customError, w)
		return
	}

	userId, errorMsg := middleware.ValdateIncomingToken(token)
	if errorMsg != "" {
		customError.S = errorMsg
		helper.GetError(&customError, w)
		return
	}

	query := bson.M{}
	strategy_id := r.URL.Query().Get("strategy_id")

	if strategy_id != "" {
		query["strategy_id"] = strategy_id
	}
	query["user_id"] = userId
	query["collection"] = "strategy"

	eventHistories := make([]models.EventHistory, 0)

	// bson.M{},  we passed empty filter. So we want to get all data.
	cur, err := eventHistoryStrategyCollection.Find(context.TODO(), query)

	if err != nil {
		helper.GetError(err, w)
		return
	}
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var eventHistory models.EventHistory
		// & character returns the memory address of the following variable.
		err := cur.Decode(&eventHistory) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// add item our array
		eventHistories = append(eventHistories, eventHistory)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	query["collection"] = "deal"
	eventHistoriesDeal := make([]models.DealEventHistory, 0)

	cur, err = eventHistoryStrategyCollection.Find(context.TODO(), query)

	if err != nil {
		helper.GetError(err, w)
		return
	}
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var eventHistoryDeal models.DealEventHistory
		// & character returns the memory address of the following variable.
		err := cur.Decode(&eventHistoryDeal) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// add item our array
		eventHistoriesDeal = append(eventHistoriesDeal, eventHistoryDeal)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(models.EventHistoryResponse{
		Strategy: &eventHistories,
		Deal:     &eventHistoriesDeal,
	})

}

var eventHistoryExchangeCollection *mongo.Collection
var eventHistoryStrategyCollection *mongo.Collection

func init() {
	//Connect to mongoDB with helper class
	eventHistoryExchangeCollection, eventHistoryStrategyCollection = helper.ConnectDB()
}

func main() {
	r := mux.NewRouter()

	// arrange our route
	r.HandleFunc("/EventHistoryService/api/v1/strategies", getEventHistoryStrategy).Methods("GET")

	// set our port address
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
