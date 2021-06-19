package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/price-service/helper"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/price-service/middleware"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/price-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getPositions(w http.ResponseWriter, r *http.Request) {
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

	var params = mux.Vars(r)
	if params["exchangeId"] == "" {
		customError.S = "Exchange Id cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	exchangeId, err := primitive.ObjectIDFromHex(params["exchangeId"])
	if err != nil {
		helper.GetError(err, w)
		return
	}
	var exchange models.Exchange
	resultReadExchange := exchangeCollection.FindOne(context.TODO(), bson.M{"_id": exchangeId, "user_id": userId})

	if err := resultReadExchange.Decode(&exchange); err != nil {
		customError.S = fmt.Sprintf("Could not find Exchange with Object Id %s: %v", exchangeId, err)
		helper.GetError(&customError, w)
		return
	}

	var positions []models.Position

	if exchange.SelectedExchange == "Alpaca" {
		os.Setenv(common.EnvApiKeyID, exchange.ApiKey)
		os.Setenv(common.EnvApiSecretKey, exchange.ApiSecret)
		if exchange.ExchangeType == "paper_trading" {
			alpaca.SetBaseUrl("https://paper-api.alpaca.markets")
		} else if exchange.ExchangeType == "live_trading" {
			alpaca.SetBaseUrl("https://api.alpaca.markets")
		}

		alpacaClient := alpaca.NewClient(common.Credentials())
		alpacaPositions, err := alpacaClient.ListPositions()
		if err != nil {
			helper.GetError(err, w)
			return
		}

		jsonPositions, _ := json.Marshal(alpacaPositions)
		_ = json.Unmarshal(jsonPositions, &positions)

	}
	json.NewEncoder(w).Encode(positions)
}

func getAssets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var customError models.ErrorString

	token := r.Header.Get("token")

	if token == "" {
		customError.S = "Token cannot be empty"
		helper.GetError(&customError, w)
		return
	}

	_, errorMsg := middleware.ValdateIncomingToken(token)
	if errorMsg != "" {
		customError.S = errorMsg
		helper.GetError(&customError, w)
		return
	}
	var assets []models.Asset

	query := bson.M{}

	name := r.URL.Query().Get("name")
	if name != "" {
		query["name"] = name
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol != "" {
		query["symbol"] = symbol
	}

	// bson.M{},  we passed empty filter. So we want to get all data.
	cur, err := assetsCollection.Find(context.TODO(), query)

	if err != nil {
		helper.GetError(err, w)
		return
	}
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var asset models.Asset
		// & character returns the memory address of the following variable.
		err := cur.Decode(&asset) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// add item our array
		assets = append(assets, asset)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(assets)

}

var exchangeCollection *mongo.Collection
var assetsCollection *mongo.Collection
var priceCollection *mongo.Collection

func init() {
	//Connect to mongoDB with helper class
	exchangeCollection, assetsCollection, priceCollection = helper.ConnectDB()
}

func main() {
	//Init Router
	r := mux.NewRouter()

	// arrange our route
	r.HandleFunc("/PriceService/api/v1/positions/{exchangeId}", getPositions).Methods("GET")
	r.HandleFunc("/PriceService/api/v1/assets", getAssets).Methods("GET")

	// set our port address
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
