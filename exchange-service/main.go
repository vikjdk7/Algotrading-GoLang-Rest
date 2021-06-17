package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/gorilla/mux"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/exchange-service/helper"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/exchange-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getExchanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// we created Book array
	var exchanges []models.Exchange

	// bson.M{},  we passed empty filter. So we want to get all data.
	cur, err := exchangeCollection.Find(context.TODO(), bson.M{})

	if err != nil {
		helper.GetError(err, w)
		return
	}

	// Close the cursor once finished
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var exchange models.Exchange
		// & character returns the memory address of the following variable.
		err := cur.Decode(&exchange) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// add item our array
		exchanges = append(exchanges, exchange)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(exchanges) // encode similar to serialize process.
}

func createExchange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var exchange models.Exchange

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&exchange)

	var customError models.ErrorString

	if exchange.SelectedExchange != "Alpaca" {
		customError.S = "Selected exchange should be Alpaca"
		helper.GetError(&customError, w)
		return
	}
	if exchange.ExchangeName == "" {
		customError.S = "ExchangeName cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if exchange.ApiKey == "" {
		customError.S = "API KEY cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if exchange.ApiSecret == "" {
		customError.S = "API SECRET cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if exchange.ExchangeType != "paper_trading" && exchange.ExchangeType != "live_trading" {
		customError.S = "Invalid Exchange Type. Allowed values: [paper_trading, live_trading]"
		helper.GetError(&customError, w)
		return
	}

	os.Setenv(common.EnvApiKeyID, exchange.ApiKey)
	os.Setenv(common.EnvApiSecretKey, exchange.ApiSecret)
	if exchange.ExchangeType == "paper_trading" {
		alpaca.SetBaseUrl("https://paper-api.alpaca.markets")
	} else if exchange.ExchangeType == "live_trading" {
		alpaca.SetBaseUrl("https://api.alpaca.markets")
	}

	alpacaClient := alpaca.NewClient(common.Credentials())

	acct, err := alpacaClient.GetAccount()
	if err != nil {
		helper.GetError(err, w)
		return
	}

	var exchangeAccount models.ExchangeAccountInfo
	jsonAcc, _ := json.Marshal(acct)
	_ = json.Unmarshal(jsonAcc, &exchangeAccount)

	// insert our exchange model.
	result, err := exchangeCollection.InsertOne(context.TODO(), exchange)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	// add the id to blog
	oid := result.InsertedID.(primitive.ObjectID)
	exchange.ID = oid

	eventHistory := models.EventHistory{
		OperationType: "insert",
		Timestamp:     time.Now().Format(time.RFC3339),
		Db:            "hedgina_algobot",
		Collection:    "exchange",
		Name:          exchange.ExchangeName,
		UserId:        exchange.UserId,
		ExchangeId:    exchange.ID.Hex(),
		NewValue:      exchange,
	}

	_, errEventHistory := eventHistoryCollection.InsertOne(context.TODO(), eventHistory)
	if errEventHistory != nil {
		helper.GetError(err, w)
		return
	}

	response := models.CreateExchangeResponse{
		Exchange:    exchange,
		AccountInfo: exchangeAccount,
	}
	json.NewEncoder(w).Encode(response)
}

var exchangeCollection *mongo.Collection
var eventHistoryCollection *mongo.Collection

func init() {
	//Connect to mongoDB with helper class
	exchangeCollection, eventHistoryCollection = helper.ConnectDB()
}

func main() {
	//Init Router
	r := mux.NewRouter()

	// arrange our route
	r.HandleFunc("/ExchangeService/api/v1/exchanges", getExchanges).Methods("GET")
	//r.HandleFunc("/ExchangeService/api/v1/exchanges/{id}", getExchange).Methods("GET")
	r.HandleFunc("/ExchangeService/api/v1/exchanges", createExchange).Methods("POST")
	//r.HandleFunc("/ExchangeService/api/v1/exchanges/{id}", updateExchange).Methods("PUT")
	//r.HandleFunc("/ExchangeService/api/v1/exchanges/{id}", deleteExchange).Methods("DELETE")

	// set our port address
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}

}
