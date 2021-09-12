package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/order-service/helper"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/order-service/middleware"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/order-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func placeOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var orderRequest models.OrderRequest
	_ = json.NewDecoder(r.Body).Decode(&orderRequest)

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
	// Convert the Id string to a MongoDB ObjectId
	exchange_id, err := primitive.ObjectIDFromHex(orderRequest.ExchangeId)
	if err != nil {
		helper.GetError(err, w)
		return
	}
	resultReadExchange := exchangeCollection.FindOne(context.TODO(), bson.M{"_id": exchange_id, "user_id": userId})
	var exchangeDataRead models.Exchange

	if err := resultReadExchange.Decode(&exchangeDataRead); err != nil {
		helper.GetError(err, w)
		return
	}

	var order models.OrderMongo

	if exchangeDataRead.SelectedExchange == "Alpaca" {
		os.Setenv(common.EnvApiKeyID, exchangeDataRead.ApiKey)
		os.Setenv(common.EnvApiSecretKey, exchangeDataRead.ApiSecret)
		if exchangeDataRead.ExchangeType == "paper_trading" {
			alpaca.SetBaseUrl("https://paper-api.alpaca.markets")
		} else if exchangeDataRead.ExchangeType == "live_trading" {
			alpaca.SetBaseUrl("https://api.alpaca.markets")
		} else {
			customError.S = "Invalid exchange type. Allowed values: [paper_trading, live_trading]"
			helper.GetError(&customError, w)
			return
		}

		placeOrderRequest := alpaca.PlaceOrderRequest{}

		if orderRequest.Symbol == "" {
			customError.S = "Symbol cannot be empty or null"
			helper.GetError(&customError, w)
			return
		} else {
			assetCount, err := assetsCollection.CountDocuments(context.TODO(), bson.M{"symbol": orderRequest.Symbol})
			if err != nil {
				helper.GetError(err, w)
				return
			}
			if assetCount == 0 {
				customError.S = fmt.Sprintf("Could not find asset %s in records", orderRequest.Symbol)
				helper.GetError(&customError, w)
				return
			}
			placeOrderRequest.AssetKey = &orderRequest.Symbol
		}
		if orderRequest.Qty.Cmp(decimal.NewFromFloat(0.0)) != 1 {
			customError.S = "Quantity should be greater than 0.0"
			helper.GetError(&customError, w)
			return
		} else {
			placeOrderRequest.Qty = orderRequest.Qty
		}

		if orderRequest.Side != models.Buy && orderRequest.Side != models.Sell {
			customError.S = "Invalid Side. Allowed Values: [buy, sell]"
			helper.GetError(&customError, w)
			return
		} else {
			if orderRequest.Side == models.Buy {
				placeOrderRequest.Side = alpaca.Buy
			} else {
				placeOrderRequest.Side = alpaca.Sell
			}
		}

		if orderRequest.Type != models.Market && orderRequest.Type != models.Limit && orderRequest.Type != models.Stop && orderRequest.Type != models.StopLimit {
			customError.S = "Invalid Order Type. Allowed Values: [market, limit, stop, stop_limit]"
			helper.GetError(&customError, w)
			return
		}
		if orderRequest.Type == models.Market {
			placeOrderRequest.Type = alpaca.Market
		} else if orderRequest.Type == models.Limit {
			if orderRequest.LimitPrice.Cmp(decimal.NewFromFloat(0.0)) != 1 {
				customError.S = "Invalid value for Limit Price. Limit Price should be > 0.0"
				helper.GetError(&customError, w)
				return
			}
			placeOrderRequest.Type = alpaca.Limit
			placeOrderRequest.LimitPrice = &orderRequest.LimitPrice
		} else if orderRequest.Type == models.Stop {
			if orderRequest.StopPrice.Cmp(decimal.NewFromFloat(0.0)) != 1 {
				customError.S = "Invalid value for Stop Price. Stop Price should be > 0.0"
				helper.GetError(&customError, w)
				return
			}
			placeOrderRequest.Type = alpaca.Stop
			placeOrderRequest.StopPrice = &orderRequest.StopPrice
		} else if orderRequest.Type == models.StopLimit {
			if orderRequest.StopPrice.Cmp(decimal.NewFromFloat(0.0)) != 1 {
				customError.S = "Invalid value for Stop Price. Stop Price should be > 0.0"
				helper.GetError(&customError, w)
				return
			}
			if orderRequest.LimitPrice.Cmp(decimal.NewFromFloat(0.0)) != 1 {
				customError.S = "Invalid value for Limit Price. Limit Price should be > 0.0"
				helper.GetError(&customError, w)
				return
			}
			placeOrderRequest.Type = alpaca.StopLimit
			placeOrderRequest.StopPrice = &orderRequest.StopPrice
			placeOrderRequest.LimitPrice = &orderRequest.LimitPrice
		}

		if orderRequest.TimeInForce == models.Day {
			placeOrderRequest.TimeInForce = alpaca.Day
		} else if orderRequest.TimeInForce == models.GTC {
			placeOrderRequest.TimeInForce = alpaca.GTC
		} else if orderRequest.TimeInForce == models.OPG {
			placeOrderRequest.TimeInForce = alpaca.OPG
		} else if orderRequest.TimeInForce == models.IOC {
			placeOrderRequest.TimeInForce = alpaca.IOC
		} else if orderRequest.TimeInForce == models.FOK {
			placeOrderRequest.TimeInForce = alpaca.FOK
		} else if orderRequest.TimeInForce == models.GTX {
			placeOrderRequest.TimeInForce = alpaca.GTX
		} else if orderRequest.TimeInForce == models.GTD {
			placeOrderRequest.TimeInForce = alpaca.GTD
		} else if orderRequest.TimeInForce == models.CLS {
			placeOrderRequest.TimeInForce = alpaca.CLS
		} else {
			customError.S = "Invalid value for Time in Force. Allowed Values: [day, gtc, opg, ioc, fok, gtx, gtd, cls]"
			helper.GetError(&customError, w)
			return
		}

		alpacaClient := alpaca.NewClient(common.Credentials())

		orderPlaced, err := alpacaClient.PlaceOrder(placeOrderRequest)

		if err != nil {
			helper.GetError(err, w)
			return
		}
		jsonOrder, _ := json.Marshal(orderPlaced)
		//fmt.Println("jsonOrder: ", jsonOrder)
		_ = json.Unmarshal(jsonOrder, &order)
		//order = orderPlaced

		order.Qty, _ = orderPlaced.Qty.Float64()
		order.Notional, _ = orderPlaced.Notional.Float64()
		order.FilledQty, _ = orderPlaced.FilledQty.Float64()

		if orderPlaced.LimitPrice != nil {
			order.LimitPrice, _ = orderPlaced.LimitPrice.Float64()
		}
		if orderPlaced.FilledAvgPrice != nil {
			order.FilledAvgPrice, _ = orderPlaced.FilledAvgPrice.Float64()
		}
		if orderPlaced.StopPrice != nil {
			order.StopPrice, _ = orderPlaced.StopPrice.Float64()
		}

		order.UserId = userId
		order.ExchangeId = orderRequest.ExchangeId

		fmt.Println("Order: ", order)

		_, err = orderCollection.InsertOne(context.TODO(), order)
		if err != nil {
			helper.GetError(&customError, w)
			return
		}
	} else {
		customError.S = "Invalid exchange. Allowed value: Alpaca"
		helper.GetError(&customError, w)
		return
	}
	json.NewEncoder(w).Encode(order)
}

func getOrders(w http.ResponseWriter, r *http.Request) {
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
	cursor, err := orderCollection.Find(context.Background(), bson.M{"user_id": userId})
	if err != nil {
		helper.GetError(err, w)
		return
	}
	defer cursor.Close(context.Background())
	orders := make([]models.OrderMongo, 0)

	for cursor.Next(context.Background()) {
		var order models.OrderMongo
		err := cursor.Decode(&order) // decode similar to deserialize process.
		if err != nil {
			helper.GetError(err, w)
			return
		}
		// add item our array
		orders = append(orders, order)
	}
	json.NewEncoder(w).Encode(orders)
}

func getOrdersTotal(w http.ResponseWriter, r *http.Request) {
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

	var orderInfo models.OrderInfo

	orderCount, err := orderCollection.CountDocuments(context.TODO(), bson.M{"user_id": userId, "status": "filled"})
	if err != nil {
		helper.GetError(err, w)
		return
	}
	orderInfo.CompletedOrderCount = orderCount

	//Completed Orders Profit
	cur, err := orderCollection.Find(context.TODO(), bson.M{"user_id": userId, "status": "filled"})
	if err != nil {
		helper.GetError(err, w)
		return
	}
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var order models.OrderMongo
		err := cur.Decode(&order) // decode similar to deserialize process.
		if err != nil {
			helper.GetError(err, w)
			return
		}
		orderInfo.OrderTotalAmount += (order.FilledQty * order.FilledAvgPrice)
	}

	//Completed Deals Profit
	statuses := [2]string{"completed", "cancelled"}
	cur, err = dealsCollection.Find(context.TODO(), bson.M{"user_id": userId, "status": bson.M{"$in": statuses}})
	if err != nil {
		helper.GetError(err, w)
		return
	}
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var deal models.Deal
		err := cur.Decode(&deal) // decode similar to deserialize process.
		if err != nil {
			helper.GetError(err, w)
			return
		}
		orderInfo.CompletedDealsProfit += deal.ProfitValue
	}

	json.NewEncoder(w).Encode(orderInfo)
}

var exchangeCollection *mongo.Collection
var orderCollection *mongo.Collection
var assetsCollection *mongo.Collection
var dealsCollection *mongo.Collection

func init() {
	//Connect to mongoDB with helper class
	exchangeCollection, orderCollection, assetsCollection, dealsCollection = helper.ConnectDB()
}

func main() {
	//Init Router
	r := mux.NewRouter()

	// arrange our route
	r.HandleFunc("/OrderService/api/v1/orders", placeOrder).Methods("POST")
	r.HandleFunc("/OrderService/api/v1/orders", getOrders).Methods("GET")
	r.HandleFunc("/OrderService/api/v1/orders/total", getOrdersTotal).Methods("GET")

	// set our port address
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
