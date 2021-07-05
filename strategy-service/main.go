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
	"github.com/vikjdk7/Algotrading-GoLang-Rest/strategy-service/helper"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/strategy-service/middleware"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/strategy-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func createStrategy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var strategy models.Strategy

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&strategy)

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

	if strategy.StrategyName == "" {
		customError.S = "Strategy Name cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if strategy.SelectedExchange == "" {
		customError.S = "Selected Exchange cannot be empty"
		helper.GetError(&customError, w)
		return
	} else {
		exchangeId, _ := primitive.ObjectIDFromHex(strategy.SelectedExchange)
		exchangeCount, err := exchangeCollection.CountDocuments(context.TODO(), bson.M{"_id": exchangeId, "user_id": userId, "active": true})
		if err != nil {
			helper.GetError(err, w)
			return
		}
		if exchangeCount == 0 {
			customError.S = fmt.Sprintf("Could not find an active exchange with id %s", strategy.SelectedExchange)
			helper.GetError(&customError, w)
			return
		}
	}
	strategy.StrategyType = "long"
	strategy.StartOrderType = "market"
	strategy.DealStartCondition = "Open new trade asap"
	strategy.Status = "stopped"
	strategy.Version = 1

	if strategy.BaseOrderSize == 0.0 {
		customError.S = "Base Order Size cannot be empty or 0.0"
		helper.GetError(&customError, w)
		return
	}
	if strategy.SafetyOrderSize == 0.0 {
		customError.S = "Safety Order Size cannot be empty or 0.0"
		helper.GetError(&customError, w)
		return
	}
	if strategy.MaxSafetyTradeCount == 0 {
		customError.S = "Max Safety Trade Count cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if strategy.MaxActiveSafetyTradeCount == 0 {
		customError.S = "Max Active Safety Trade Count cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if strategy.PriceDevation == "" {
		customError.S = "Price Deviation cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if strategy.SafetyOrderVolumeScale == 0.0 {
		customError.S = "Safety Order Volume Scale cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if strategy.SafetyOrderStepScale == 0.0 {
		customError.S = "Safety Order Step Scale cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if strategy.TargetProfit == "" {
		customError.S = "Target Profit cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	if strategy.StopLossPercent == "" {
		customError.S = "Stop Loss Percent cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	strategy.UserId = userId
	if strategy.Stock == nil {
		customError.S = "Stocks cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	for _, v := range strategy.Stock {
		assetCount, err := assetsCollection.CountDocuments(context.TODO(), bson.M{"symbol": v.StockName})
		if err != nil {
			helper.GetError(err, w)
			return
		}
		if assetCount == 0 {
			customError.S = fmt.Sprintf("Could not find asset %s in records", v.StockName)
			helper.GetError(&customError, w)
			return
		}
	}
	// insert our strategy model.
	result, err := strategyCollection.InsertOne(context.TODO(), strategy)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	strategy.Id = result.InsertedID.(primitive.ObjectID)

	eventData := models.EventHistory{
		OperationType: "insert",
		Timestamp:     time.Now().Format(time.RFC3339),
		Db:            "hedgina_algobot",
		Collection:    "strategy",
		Name:          strategy.StrategyName,
		UserId:        strategy.UserId,
		StrategyId:    strategy.Id.Hex(),
		NewValue:      strategy,
	}
	_, errEventHistory := eventHistoryCollection.InsertOne(context.TODO(), eventData)
	if errEventHistory != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(strategy)

}

func getStrategies(w http.ResponseWriter, r *http.Request) {
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

	// create Strategy array
	var strategies []models.Strategy

	cur, err := strategyCollection.Find(context.TODO(), bson.M{"user_id": userId})

	if err != nil {
		helper.GetError(err, w)
		return
	}

	// Close the cursor once finished
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var strategy models.Strategy
		// & character returns the memory address of the following variable.
		err := cur.Decode(&strategy) // decode similar to deserialize process.
		if err != nil {
			helper.GetError(err, w)
			return
		}

		// add item our array
		strategies = append(strategies, strategy)
	}

	if err := cur.Err(); err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(strategies)
}

func getStrategy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var strategy models.Strategy

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

	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		helper.GetError(err, w)
		return
	}
	err = strategyCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId}).Decode(&strategy)
	if err != nil {
		helper.GetError(err, w)
		return
	}
	json.NewEncoder(w).Encode(strategy)

}

func updateStrategy(w http.ResponseWriter, r *http.Request) {
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

	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		helper.GetError(err, w)
		return
	}

	var strategy models.Strategy
	_ = json.NewDecoder(r.Body).Decode(&strategy)

	resultReadStrategy := strategyCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId})
	// Create an empty ExchangeItem to write our decode result to
	var dataResultReadStrategy models.StrategyRevision
	// decode and write to data
	if err := resultReadStrategy.Decode(&dataResultReadStrategy); err != nil {
		customError.S = fmt.Sprintf("Could not find Strategy with Object Id %s: %v", id, err)
		helper.GetError(&customError, w)
		return
	}

	dataResultReadStrategy.StrategyId = id.Hex()
	dataResultReadStrategy.Id = primitive.NewObjectID()

	_, err = strategy_revisionsCollection.InsertOne(context.TODO(), dataResultReadStrategy)
	if err != nil {
		helper.GetError(err, w)
		return
	}

	var oldValues models.Strategy
	var newValues models.Strategy

	// Convert the data to be updated into an unordered Bson document
	update := bson.M{}

	if strategy.StrategyName != "" {
		update["strategy_name"] = strategy.StrategyName
		oldValues.StrategyName = dataResultReadStrategy.StrategyName
		newValues.StrategyName = strategy.StrategyName
	}
	if strategy.SelectedExchange != "" {
		exchangeId, _ := primitive.ObjectIDFromHex(strategy.SelectedExchange)
		exchangeCount, err := exchangeCollection.CountDocuments(context.TODO(), bson.M{"_id": exchangeId, "user_id": userId, "active": true})
		if err != nil {
			helper.GetError(err, w)
			return
		}
		if exchangeCount == 0 {
			customError.S = fmt.Sprintf("Could not find an active exchange with id %s", strategy.SelectedExchange)
			helper.GetError(&customError, w)
			return
		}
		update["selected_exchange"] = strategy.SelectedExchange
		oldValues.SelectedExchange = dataResultReadStrategy.SelectedExchange
		newValues.SelectedExchange = strategy.SelectedExchange
	}
	if strategy.BaseOrderSize != 0.0 {
		update["base_order_size"] = strategy.BaseOrderSize
		oldValues.BaseOrderSize = dataResultReadStrategy.BaseOrderSize
		newValues.BaseOrderSize = strategy.BaseOrderSize
	}
	if strategy.SafetyOrderSize != 0.0 {
		update["safety_order_size"] = strategy.SafetyOrderSize
		oldValues.SafetyOrderSize = dataResultReadStrategy.SafetyOrderSize
		newValues.SafetyOrderSize = strategy.SafetyOrderSize
	}
	if strategy.MaxSafetyTradeCount != 0 {
		update["max_safety_trade_count"] = strategy.MaxSafetyTradeCount
		oldValues.MaxSafetyTradeCount = dataResultReadStrategy.MaxSafetyTradeCount
		newValues.MaxSafetyTradeCount = strategy.MaxSafetyTradeCount
	}
	if strategy.MaxActiveSafetyTradeCount != 0 {
		update["max_active_safety_trade_count"] = strategy.MaxActiveSafetyTradeCount
		oldValues.MaxActiveSafetyTradeCount = dataResultReadStrategy.MaxActiveSafetyTradeCount
		newValues.MaxActiveSafetyTradeCount = strategy.MaxActiveSafetyTradeCount
	}
	if strategy.PriceDevation != "" {
		update["price_devation"] = strategy.PriceDevation
		oldValues.PriceDevation = dataResultReadStrategy.PriceDevation
		newValues.PriceDevation = strategy.PriceDevation
	}
	if strategy.SafetyOrderVolumeScale != 0.0 {
		update["safety_order_volume_scale"] = strategy.SafetyOrderVolumeScale
		oldValues.SafetyOrderVolumeScale = dataResultReadStrategy.SafetyOrderVolumeScale
		newValues.SafetyOrderVolumeScale = strategy.SafetyOrderVolumeScale
	}
	if strategy.SafetyOrderStepScale != 0.0 {
		update["safety_order_step_scale"] = strategy.SafetyOrderStepScale
		oldValues.SafetyOrderStepScale = dataResultReadStrategy.SafetyOrderStepScale
		newValues.SafetyOrderStepScale = strategy.SafetyOrderStepScale
	}
	if strategy.TargetProfit == "" {
		update["target_profit"] = strategy.TargetProfit
		oldValues.TargetProfit = dataResultReadStrategy.TargetProfit
		newValues.TargetProfit = strategy.TargetProfit
	}
	if strategy.StopLossPercent == "" {
		update["stop_loss_percent"] = strategy.StopLossPercent
		oldValues.StopLossPercent = dataResultReadStrategy.StopLossPercent
		newValues.StopLossPercent = strategy.StopLossPercent
	}
	if strategy.Stock != nil {

		for _, v := range strategy.Stock {
			assetCount, err := assetsCollection.CountDocuments(context.TODO(), bson.M{"symbol": v.StockName})
			if err != nil {
				helper.GetError(err, w)
				return
			}
			if assetCount == 0 {
				customError.S = fmt.Sprintf("Could not find asset %s in records", v.StockName)
				helper.GetError(&customError, w)
				return
			}
		}
		update["stock"] = strategy.Stock
		oldValues.Stock = dataResultReadStrategy.Stock
		newValues.Stock = strategy.Stock
	}
	update["version"] = dataResultReadStrategy.Version + 1
	oldValues.Version = dataResultReadStrategy.Version
	newValues.Version = dataResultReadStrategy.Version + 1

	filter := bson.M{"_id": id}
	//fmt.Println(update)
	// Result is the BSON encoded result
	// To return the updated document instead of original we have to add options.
	result := strategyCollection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(1))

	var decoded models.Strategy
	err = result.Decode(&decoded)
	if err != nil {
		helper.GetError(err, w)
		return
	}

	strategy.Id = id
	eventData := models.EventHistory{
		OperationType: "update",
		Timestamp:     time.Now().Format(time.RFC3339),
		Db:            "hedgina_algobot",
		Collection:    "strategy",
		Name:          dataResultReadStrategy.StrategyName,
		UserId:        dataResultReadStrategy.UserId,
		StrategyId:    strategy.Id.Hex(),
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

func deleteStrategy(w http.ResponseWriter, r *http.Request) {
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
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		helper.GetError(err, w)
		return
	}

	dealsCount, err := dealsCollection.CountDocuments(context.TODO(), bson.M{"strategy_id": params["id"], "status": "running"})
	if err != nil {
		helper.GetError(err, w)
		return
	}
	if dealsCount < 1 {
		resultReadStrategy := strategyCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId})
		var dataResultReadStrategy models.Strategy
		// decode and write to data
		if err := resultReadStrategy.Decode(&dataResultReadStrategy); err != nil {
			customError.S = fmt.Sprintf("Could not find Strategy with Object Id %s: %v", id, err)
			helper.GetError(&customError, w)
			return
		}
		deleteResult, err := strategyCollection.DeleteOne(context.TODO(), bson.M{"_id": id})
		if err != nil {
			customError.S = fmt.Sprintf("Could not find/delete strategy with id %s: %v", id, err)
			helper.GetError(&customError, w)
			return
		}
		_, err = strategy_revisionsCollection.DeleteMany(context.TODO(), bson.M{"strategy_id": id.Hex()})
		if err != nil {
			customError.S = fmt.Sprintf("Could not delete from strategy_revisionsdb with id %s: %v", params["id"], err)
			helper.GetError(&customError, w)
			return
		}
		eventData := models.EventHistory{
			OperationType: "delete",
			Timestamp:     time.Now().Format(time.RFC3339),
			Db:            "hedgina_algobot",
			Collection:    "strategy",
			Name:          dataResultReadStrategy.StrategyName,
			UserId:        dataResultReadStrategy.UserId,
			StrategyId:    id.Hex(),
			OldValue:      dataResultReadStrategy,
		}
		_, errEventHistory := eventHistoryCollection.InsertOne(context.TODO(), eventData)
		if errEventHistory != nil {
			helper.GetError(err, w)
			return
		}
		json.NewEncoder(w).Encode(deleteResult)
	} else {
		customError.S = fmt.Sprintf("Cannot delete strategy with %d running deal(s)", dealsCount)
		helper.GetError(&customError, w)
		return
	}

}

func HasStock(list []*models.Stock, a string) bool {
	for _, b := range list {
		if b.StockName == a {
			return true
		}
	}
	return false
}

func createDeal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var dealRequest models.DealRequest

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&dealRequest)

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

	var strategy models.Strategy
	if dealRequest.StrategyId == "" {
		customError.S = "Strategy Id cannot be empty"
		helper.GetError(&customError, w)
		return
	}
	strategyId, err := primitive.ObjectIDFromHex(dealRequest.StrategyId)
	if err != nil {
		customError.S = "Invalid Strategy Id. Cannot convert Strategy Id to Primitive Object Id."
		helper.GetError(&customError, w)
		return
	}
	err = strategyCollection.FindOne(context.TODO(), bson.M{"_id": strategyId, "user_id": userId}).Decode(&strategy)
	if err != nil {
		customError.S = "Invalid Strategy Id. Cannot find strategy for the user."
		helper.GetError(&customError, w)
		return
	}

	var insertDeal []interface{}
	var insertDealHistory []interface{}
	for _, v := range *dealRequest.Stock {

		if !HasStock(strategy.Stock, v.StockName) {
			customError.S = fmt.Sprintf("Invalid Asset %s. Asset not part of Strategy", v.StockName)
			helper.GetError(&customError, w)
			return
		}
		dealId := primitive.NewObjectID()
		deal := bson.M{
			"_id":                           dealId,
			"strategy_id":                   dealRequest.StrategyId,
			"user_id":                       userId,
			"stock":                         v.StockName,
			"status":                        "running",
			"max_active_safety_trade_count": strategy.MaxActiveSafetyTradeCount,
			"max_safety_trade_count":        strategy.MaxSafetyTradeCount,
		}
		insertDeal = append(insertDeal, deal)
		dealHistory := bson.M{
			"operation_type": "insert",
			"timestamp":      time.Now().Format(time.RFC3339),
			"db":             "hedgina_algobot",
			"collection":     "deal",
			"name":           strategy.StrategyName,
			"user_id":        userId,
			"deal_id":        dealId,
			"strategy_id":    dealRequest.StrategyId,
			"new_value":      deal,
		}
		insertDealHistory = append(insertDealHistory, dealHistory)
	}

	insertManyResult, err := dealsCollection.InsertMany(context.TODO(), insertDeal)
	if err != nil {
		customError.S = "Could not create Deals."
		helper.GetError(&customError, w)
		return
	}
	fmt.Print("Inserted Deal ID's: ")
	fmt.Println(insertManyResult.InsertedIDs)

	insertManyDealHistoryResult, err := eventHistoryCollection.InsertMany(context.TODO(), insertDealHistory)
	if err != nil {
		customError.S = "Could not create Deals History."
		helper.GetError(&customError, w)
		return
	}
	fmt.Print("Inserted Deal History ID's: ")
	fmt.Println(insertManyDealHistoryResult.InsertedIDs)

	err = strategyCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": strategyId}, bson.M{"$set": bson.M{"status": "running"}}, options.FindOneAndUpdate().SetReturnDocument(1)).Decode(&strategy)
	if err != nil {
		customError.S = "Could not update Starategy Status."
		helper.GetError(&customError, w)
		return
	}
	strategyEventData := models.EventHistory{
		OperationType: "update",
		Timestamp:     time.Now().Format(time.RFC3339),
		Db:            "hedgina_algobot",
		Collection:    "strategy",
		Name:          strategy.StrategyName,
		UserId:        strategy.UserId,
		StrategyId:    dealRequest.StrategyId,
		OldValue: models.Strategy{
			Status: strategy.Status,
		},
		NewValue: models.Strategy{
			Status: "running",
		},
	}
	_, errStrategyEventHistory := eventHistoryCollection.InsertOne(context.TODO(), strategyEventData)
	if errStrategyEventHistory != nil {
		customError.S = "Could not create strategy event history."
		helper.GetError(&customError, w)
		return
	}

	json.NewEncoder(w).Encode(insertManyResult.InsertedIDs)

}

func getDeals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var customError models.ErrorString
	var params = mux.Vars(r)

	if params["id"] == "" {
		customError.S = "Strategy Id cannot be empty"
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

	// create Deals array
	var deals []models.Deal

	query := bson.M{}
	query["user_id"] = userId
	query["strategy_id"] = params["id"]
	status := r.URL.Query().Get("status")
	if status != "" {
		if status != "running" {
			customError.S = "Invalid query paramter status. Allowed values: [running, completed]"
			helper.GetError(&customError, w)
			return
		}
		query["status"] = status
	}

	cur, err := dealsCollection.Find(context.TODO(), query)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	// Close the cursor once finished
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var deal models.Deal
		// & character returns the memory address of the following variable.
		err := cur.Decode(&deal) // decode similar to deserialize process.
		if err != nil {
			helper.GetError(err, w)
			return
		}

		// add item our array
		deals = append(deals, deal)
	}

	if err := cur.Err(); err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(deals)
}

func getAccountInfo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var customError models.ErrorString
	var params = mux.Vars(r)
	if params["id"] == "" {
		customError.S = "Exchange Id cannot be empty"
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

	// create model
	var exchange models.Exchange
	var accountinfo models.AccountInfo

	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		customError.S = "Invalid Exchange Id. Cannot convert Exchange Id to Primitive Object Id."
		helper.GetError(&customError, w)
		return
	}
	err = exchangeCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId}).Decode(&exchange)
	if err != nil {
		customError.S = "Invalid Exchange Id. Cannot find exchange for the user."
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
	accountinfo.Balance = acct.PortfolioValue.Add(acct.Cash)
	accountinfo.MaxAmtStrategyUsage = acct.BuyingPower
	accountinfo.MaxSafetyOrderPriceDeviation = "50"
	accountinfo.AvailableBalance = 1.81
	json.NewEncoder(w).Encode(accountinfo)
}

var strategyCollection *mongo.Collection
var eventHistoryCollection *mongo.Collection
var strategy_revisionsCollection *mongo.Collection
var dealsCollection *mongo.Collection
var exchangeCollection *mongo.Collection
var assetsCollection *mongo.Collection

func init() {
	//Connect to mongoDB with helper class
	strategyCollection, eventHistoryCollection, strategy_revisionsCollection, dealsCollection, exchangeCollection, assetsCollection = helper.ConnectDB()
}

func main() {
	//Init Router
	r := mux.NewRouter()

	r.HandleFunc("/StrategyService/api/v1/strategies", createStrategy).Methods("POST")
	r.HandleFunc("/StrategyService/api/v1/strategies", getStrategies).Methods("GET")
	r.HandleFunc("/StrategyService/api/v1/strategies/{id}", getStrategy).Methods("GET")
	r.HandleFunc("/StrategyService/api/v1/strategies/{id}", updateStrategy).Methods("PUT")
	r.HandleFunc("/StrategyService/api/v1/strategies/{id}", deleteStrategy).Methods("DELETE")

	r.HandleFunc("/StrategyService/api/v1/deals", createDeal).Methods("POST")
	r.HandleFunc("/StrategyService/api/v1/deals/{id}", getDeals).Methods("GET")

	r.HandleFunc("/StrategyService/api/v1/accountinfo/{id}", getAccountInfo).Methods("GET")

	// set our port address
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
