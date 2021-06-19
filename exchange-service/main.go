package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/gorilla/mux"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/exchange-service/helper"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/exchange-service/middleware"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/exchange-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getExchanges(w http.ResponseWriter, r *http.Request) {
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
	//fmt.Println(userId)

	// we created Exchange array
	var exchanges []models.Exchange

	// bson.M{},  we passed empty filter. So we want to get all data.
	cur, err := exchangeCollection.Find(context.TODO(), bson.M{"user_id": userId})

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

func getExchange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var exchange models.Exchange
	var customError models.ErrorString

	var params = mux.Vars(r)
	if params["id"] == "" {
		customError.S = "Id cannot be empty"
		helper.GetError(&customError, w)
		return
	}
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

	id, _ := primitive.ObjectIDFromHex(params["id"])

	err := exchangeCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId}).Decode(&exchange)
	if err != nil {
		helper.GetError(err, w)
		return
	}
	json.NewEncoder(w).Encode(exchange)
}

func createExchange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var exchange models.Exchange

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&exchange)

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

	exchange.UserId = userId

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

func updateExchange(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var params = mux.Vars(r)

	var customError models.ErrorString

	if params["id"] == "" {
		customError.S = "Id cannot be empty"
		helper.GetError(&customError, w)
		return
	}
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

	id, _ := primitive.ObjectIDFromHex(params["id"])

	var exchange models.Exchange
	_ = json.NewDecoder(r.Body).Decode(&exchange)

	resultReadExchange := exchangeCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId})
	// Create an empty ExchangeItem to write our decode result to
	var dataResultReadExchange models.Exchange
	// decode and write to data
	if err := resultReadExchange.Decode(&dataResultReadExchange); err != nil {
		customError.S = fmt.Sprintf("Could not find Exchange with Object Id %s: %v", id, err)
		helper.GetError(&customError, w)
		return
	}

	var oldValues models.Exchange
	var newValues models.Exchange

	// Convert the data to be updated into an unordered Bson document
	update := bson.M{}

	if exchange.SelectedExchange != "" {
		if exchange.SelectedExchange != "Alpaca" {
			customError.S = "Selected exchange should be Alpaca"
			helper.GetError(&customError, w)
			return
		}
		update["selected_exchange"] = exchange.SelectedExchange
		oldValues.SelectedExchange = dataResultReadExchange.SelectedExchange
		newValues.SelectedExchange = exchange.SelectedExchange
	}
	if exchange.ExchangeName != "" {
		update["exchange_name"] = exchange.ExchangeName
		oldValues.ExchangeName = dataResultReadExchange.ExchangeName
		newValues.ExchangeName = exchange.ExchangeName
	}
	if exchange.ExchangeType != "" || exchange.ApiKey != "" || exchange.ApiSecret != "" {
		if exchange.ExchangeType != "" {
			if exchange.ExchangeType != "paper_trading" && exchange.ExchangeType != "live_trading" {
				customError.S = "Invalid Exchange Type. Allowed values: [paper_trading, live_trading]"
				helper.GetError(&customError, w)
				return
			}
			update["exchange_type"] = exchange.ExchangeType
			oldValues.ExchangeType = dataResultReadExchange.ExchangeType
			newValues.ExchangeType = exchange.ExchangeType
			if exchange.ExchangeType == "paper_trading" {
				alpaca.SetBaseUrl("https://paper-api.alpaca.markets")
			} else if exchange.ExchangeType == "live_trading" {
				alpaca.SetBaseUrl("https://api.alpaca.markets")
			}
		}

		os.Setenv(common.EnvApiKeyID, exchange.ApiKey)
		os.Setenv(common.EnvApiSecretKey, exchange.ApiSecret)

		alpacaClient := alpaca.NewClient(common.Credentials())
		_, err := alpacaClient.GetAccount()
		if err != nil {
			helper.GetError(err, w)
			return
		}

		if exchange.ApiKey != "" {
			update["api_key"] = exchange.ApiKey
			oldValues.ApiKey = dataResultReadExchange.ApiKey
			newValues.ApiKey = exchange.ApiKey
		}
		if exchange.ApiSecret != "" {
			update["api_secret"] = exchange.ApiSecret
			oldValues.ApiSecret = dataResultReadExchange.ApiSecret
			newValues.ApiSecret = exchange.ApiSecret
		}

	}
	filter := bson.M{"_id": id}
	//fmt.Println(update)
	// Result is the BSON encoded result
	// To return the updated document instead of original we have to add options.
	result := exchangeCollection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(1))

	// Decode result and write it to 'decoded'
	var decoded models.Exchange
	err := result.Decode(&decoded)
	if err != nil {
		helper.GetError(err, w)
		return
	}

	exchange.ID = id
	eventData := models.EventHistory{
		OperationType: "update",
		Timestamp:     time.Now().Format(time.RFC3339),
		Db:            "hedgina_algobot",
		Collection:    "exchange",
		Name:          dataResultReadExchange.ExchangeName,
		UserId:        dataResultReadExchange.UserId,
		ExchangeId:    exchange.ID.Hex(),
		OldValue:      oldValues,
		NewValue:      newValues,
	}
	_, errEventHistory := eventHistoryCollection.InsertOne(context.TODO(), eventData)
	if errEventHistory != nil {
		helper.GetError(errEventHistory, w)
		return
	}
	json.NewEncoder(w).Encode(decoded)
}

func deleteExchange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var params = mux.Vars(r)

	var customError models.ErrorString

	if params["id"] == "" {
		customError.S = "Id cannot be empty"
		helper.GetError(&customError, w)
		return
	}
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

	id, _ := primitive.ObjectIDFromHex(params["id"])

	resultReadExchange := exchangeCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId})
	// Create an empty ExchangeItem to write our decode result to
	var dataResultReadExchange models.Exchange
	// decode and write to data
	if err := resultReadExchange.Decode(&dataResultReadExchange); err != nil {
		customError.S = fmt.Sprintf("Could not find Exchange with Object Id %s: %v", id, err)
		helper.GetError(&customError, w)
		return
	}

	// DeleteOne returns DeleteResult which is a struct containing the amount of deleted docs (in this case only 1 always)
	// So we return a boolean instead
	deleteResult, err := exchangeCollection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		customError.S = fmt.Sprintf("Could not find/delete exchange with id %s: %v", id, err)
		helper.GetError(&customError, w)
		return
	}

	eventData := models.EventHistory{
		OperationType: "delete",
		Timestamp:     time.Now().Format(time.RFC3339),
		Db:            "hedgina_algobot",
		Collection:    "exchange",
		Name:          dataResultReadExchange.ExchangeName,
		UserId:        dataResultReadExchange.UserId,
		ExchangeId:    id.Hex(),
		OldValue:      dataResultReadExchange,
	}
	_, errEventHistory := eventHistoryCollection.InsertOne(context.TODO(), eventData)
	if errEventHistory != nil {
		helper.GetError(errEventHistory, w)
		return
	}

	json.NewEncoder(w).Encode(deleteResult)
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
	r.HandleFunc("/ExchangeService/api/v1/exchanges/{id}", getExchange).Methods("GET")
	r.HandleFunc("/ExchangeService/api/v1/exchanges", createExchange).Methods("POST")
	r.HandleFunc("/ExchangeService/api/v1/exchanges/{id}", updateExchange).Methods("PUT")
	r.HandleFunc("/ExchangeService/api/v1/exchanges/{id}", deleteExchange).Methods("DELETE")

	// set our port address
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}

}
