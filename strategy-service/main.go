package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/gorilla/mux"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/strategy-service/algobot"
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
	} else {
		strategyCount, err := strategyCollection.CountDocuments(context.TODO(), bson.M{"strategy_name": strategy.StrategyName, "user_id": userId})
		if err != nil {
			helper.GetError(err, w)
			return
		}
		if strategyCount > 0 {
			customError.S = fmt.Sprintf("Duplicate Name. Strategy with name %s already exixts", strategy.StrategyName)
			helper.GetError(&customError, w)
			return
		}
	}
	if strategy.SelectedExchange == "" {
		customError.S = "Selected Exchange cannot be empty"
		helper.GetError(&customError, w)
		return
	} else {
		if strategy.SelectedExchangeName == "" {
			customError.S = "Selected Exchange Name cannot be empty"
			helper.GetError(&customError, w)
			return
		}
		exchangeId, _ := primitive.ObjectIDFromHex(strategy.SelectedExchange)
		exchangeCount, err := exchangeCollection.CountDocuments(context.TODO(), bson.M{"_id": exchangeId, "user_id": userId, "active": true, "exchange_name": strategy.SelectedExchangeName})
		if err != nil {
			helper.GetError(err, w)
			return
		}
		if exchangeCount == 0 {
			customError.S = fmt.Sprintf("Cannot find an active exchange for the user with id %s and Name %s", strategy.SelectedExchange, strategy.SelectedExchangeName)
			helper.GetError(&customError, w)
			return
		}
	}
	strategy.StrategyType = "long"
	strategy.StartOrderType = "market"
	strategy.DealStartCondition = "Open new trade asap"
	strategy.Status = "running"
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

	strategy.Id = primitive.NewObjectID()
	strategy.TotalDeals = int64(len(strategy.Stock))

	dealsArray, errMsg := createDeal(strategy)

	if errMsg != "" {
		customError.S = errMsg
		helper.GetError(&customError, w)
		return
	}

	//Waitgroup
	wg := &sync.WaitGroup{}

	// Get Exchange Data for Creating Deals
	exchange_id, _ := primitive.ObjectIDFromHex(strategy.SelectedExchange)
	var dataResultReadExchange models.Exchange
	resultReadExchange := exchangeCollection.FindOne(context.TODO(), bson.M{"_id": exchange_id, "user_id": strategy.UserId})
	if err := resultReadExchange.Decode(&dataResultReadExchange); err != nil {
		customError.S = fmt.Sprintf("Could not find Exchange with Id %s: %v", exchange_id, err)
		helper.GetError(&customError, w)
		return
	}
	var exchange_url string
	if dataResultReadExchange.ExchangeType == "paper_trading" {
		exchange_url = "https://paper-api.alpaca.markets"
	} else {
		exchange_url = "https://api.alpaca.markets"
	}

	var active_deals int64 = 0
	for _, v := range dealsArray {
		wg.Add(1)
		errMsg := algobot.StartBot(strategy, v.DealId, v.Asset, dataResultReadExchange.ApiKey, dataResultReadExchange.ApiSecret, exchange_url, wg)
		if errMsg != "" {
			fmt.Println(fmt.Sprintf("Could not start bot for deal %s. Error: %s", v.DealId, errMsg))
			//customError.S = fmt.Sprintf("Could not start bot for deal %s. Error: %s", v.DealId, errMsg)
			deal_id, _ := primitive.ObjectIDFromHex(v.DealId)
			_, err := dealsCollection.DeleteOne(context.TODO(), bson.M{"_id": deal_id})
			if err != nil {
				fmt.Sprintf("Could not find/delete deal with id %s: %v", deal_id, err)
			}
			//helper.GetError(&customError, w)
			//return
		} else {
			active_deals++
		}
	}
	strategy.ActiveDeals = active_deals
	// insert our strategy model.
	_, err := strategyCollection.InsertOne(context.TODO(), strategy)

	if err != nil {
		helper.GetError(err, w)
		return
	}

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

	wg.Wait()
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
	strategies := make([]models.Strategy, 0)

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
		strategyCount, err := strategyCollection.CountDocuments(context.TODO(), bson.M{"strategy_name": strategy.StrategyName, "user_id": userId})
		if err != nil {
			helper.GetError(err, w)
			return
		}
		if strategyCount > 0 && strategy.StrategyName != dataResultReadStrategy.StrategyName {
			customError.S = fmt.Sprintf("Duplicate Name. Strategy with name %s already exixts", strategy.StrategyName)
			helper.GetError(&customError, w)
			return
		}
		update["strategy_name"] = strategy.StrategyName
		oldValues.StrategyName = dataResultReadStrategy.StrategyName
		newValues.StrategyName = strategy.StrategyName
	}
	if strategy.SelectedExchange != "" {
		if strategy.SelectedExchangeName == "" {
			customError.S = "Selected Exchange Name cannot be empty"
			helper.GetError(&customError, w)
			return
		}
		exchangeId, _ := primitive.ObjectIDFromHex(strategy.SelectedExchange)
		exchangeCount, err := exchangeCollection.CountDocuments(context.TODO(), bson.M{"_id": exchangeId, "user_id": userId, "active": true, "exchange_name": strategy.SelectedExchangeName})
		if err != nil {
			helper.GetError(err, w)
			return
		}
		if exchangeCount == 0 {
			customError.S = fmt.Sprintf("Cannot find an active exchange for the user with id %s and Name %s", strategy.SelectedExchange, strategy.SelectedExchangeName)
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

		oldValues.TotalDeals = dataResultReadStrategy.TotalDeals
		newValues.TotalDeals = int64(len(strategy.Stock))
		update["total_deals"] = newValues.TotalDeals
		oldValues.ActiveDeals = dataResultReadStrategy.ActiveDeals
		var active_deals int64 = dataResultReadStrategy.ActiveDeals

		exchange_id, _ := primitive.ObjectIDFromHex(strategy.SelectedExchange)
		var dataResultReadExchange models.Exchange
		resultReadExchange := exchangeCollection.FindOne(context.TODO(), bson.M{"_id": exchange_id, "user_id": userId})
		if err := resultReadExchange.Decode(&dataResultReadExchange); err != nil {
			customError.S = fmt.Sprintf("Could not find Exchange with Id %s: %v", exchange_id, err)
			helper.GetError(&customError, w)
			return
		}
		var exchange_url string
		if dataResultReadExchange.ExchangeType == "paper_trading" {
			exchange_url = "https://paper-api.alpaca.markets"
		} else {
			exchange_url = "https://api.alpaca.markets"
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

			dealCount, err := dealsCollection.CountDocuments(context.TODO(), bson.M{"stock": v.StockName, "user_id": userId, "strategy_id": dataResultReadStrategy.StrategyId})
			if err != nil {
				helper.GetError(err, w)
				return
			}
			if dealCount == 0 {

				dealId := primitive.NewObjectID()

				max_active_safety_trade_count := dataResultReadStrategy.MaxActiveSafetyTradeCount
				if strategy.MaxActiveSafetyTradeCount != 0 {
					max_active_safety_trade_count = strategy.MaxActiveSafetyTradeCount
				}

				max_safety_trade_count := dataResultReadStrategy.MaxSafetyTradeCount

				if strategy.MaxSafetyTradeCount != 0 {
					max_safety_trade_count = strategy.MaxSafetyTradeCount
				}

				name := dataResultReadStrategy.StrategyName
				if strategy.StrategyName != "" {
					name = strategy.StrategyName
				}
				current_timestamp, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

				dealInsert := bson.M{
					"_id":                           dealId,
					"strategy_id":                   dataResultReadStrategy.StrategyId,
					"strategy_version":              dataResultReadStrategy.Version + 1,
					"user_id":                       userId,
					"stock":                         v.StockName,
					"status":                        "running",
					"max_active_safety_trade_count": max_active_safety_trade_count,
					"max_safety_trade_count":        max_safety_trade_count,
					"active_safety_order_count":     0,
					"filled_safety_order_count":     0,
					"created_at":                    current_timestamp,
					"total_order_quantity":          0,
					"profit_percentage":             "0",
					"total_buying_price":            0.0,
					"total_sell_price":              0.0,
					"target_profit":                 strategy.TargetProfit,
					"strategy_name":                 strategy.StrategyName,
					"selected_exchange":             strategy.SelectedExchange,
					"base_order_size":               strategy.BaseOrderSize,
					"safety_order_size":             strategy.SafetyOrderSize,
				}

				dealHistory := bson.M{
					"operation_type": "insert",
					"timestamp":      time.Now().Format(time.RFC3339),
					"db":             "hedgina_algobot",
					"collection":     "deal",
					"name":           name,
					"user_id":        userId,
					"deal_id":        dealId,
					"strategy_id":    dataResultReadStrategy.StrategyId,
					"new_value":      dealInsert,
				}
				wg := &sync.WaitGroup{}

				wg.Add(1)
				errMsg := algobot.StartBot(strategy, dealId.Hex(), v.StockName, dataResultReadExchange.ApiKey, dataResultReadExchange.ApiSecret, exchange_url, wg)
				if errMsg != "" {
					fmt.Println(fmt.Sprintf("Could not start bot for the deal. Error: %s", errMsg))
					//customError.S = fmt.Sprintf("Could not start bot for the deal. Error: %s", errMsg)
					//helper.GetError(&customError, w)
					//return
				} else {
					_, err := dealsCollection.InsertOne(context.TODO(), dealInsert)
					if err != nil {
						customError.S = "Could not create Deal."
						helper.GetError(&customError, w)
						return
					}

					_, err = eventHistoryCollection.InsertOne(context.TODO(), dealHistory)
					if err != nil {
						customError.S = "Could not create Deal History."
						helper.GetError(&customError, w)
						return
					}
					active_deals++
				}
				wg.Wait()

				update["status"] = "running"

			}
		}
		newValues.ActiveDeals = active_deals
		update["active_deals"] = active_deals
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

func createDeal(strategy models.Strategy) (dealsArray []models.DealJson, msg string) {

	var insertDeal []interface{}
	var insertDealHistory []interface{}
	dealsArray = make([]models.DealJson, 0)
	var dealJson models.DealJson
	for _, v := range strategy.Stock {
		dealId := primitive.NewObjectID()
		current_timestamp, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		dealJson.DealId = dealId.Hex()
		dealJson.Asset = v.StockName
		dealsArray = append(dealsArray, dealJson)
		deal := bson.M{
			"_id":                           dealId,
			"strategy_id":                   strategy.Id.Hex(),
			"strategy_version":              1,
			"user_id":                       strategy.UserId,
			"stock":                         v.StockName,
			"status":                        "running",
			"max_active_safety_trade_count": strategy.MaxActiveSafetyTradeCount,
			"max_safety_trade_count":        strategy.MaxSafetyTradeCount,
			"active_safety_order_count":     0,
			"filled_safety_order_count":     0,
			"created_at":                    current_timestamp,
			"total_order_quantity":          0,
			"profit_percentage":             "0",
			"total_buying_price":            0.0,
			"total_sell_price":              0.0,
			"target_profit":                 strategy.TargetProfit,
			"strategy_name":                 strategy.StrategyName,
			"selected_exchange":             strategy.SelectedExchange,
			"base_order_size":               strategy.BaseOrderSize,
			"safety_order_size":             strategy.SafetyOrderSize,
		}
		insertDeal = append(insertDeal, deal)
		dealHistory := bson.M{
			"operation_type": "insert",
			"timestamp":      time.Now().Format(time.RFC3339),
			"db":             "hedgina_algobot",
			"collection":     "deal",
			"name":           strategy.StrategyName,
			"user_id":        strategy.UserId,
			"deal_id":        dealId,
			"strategy_id":    strategy.Id.Hex(),
			"new_value":      deal,
		}
		insertDealHistory = append(insertDealHistory, dealHistory)
	}

	insertManyResult, err := dealsCollection.InsertMany(context.TODO(), insertDeal)
	if err != nil {
		return nil, "Could not create Deals."
	}
	fmt.Print("Inserted Deal ID's: ")
	fmt.Println(insertManyResult.InsertedIDs)

	insertManyDealHistoryResult, err := eventHistoryCollection.InsertMany(context.TODO(), insertDealHistory)
	if err != nil {
		return nil, "Could not create Deals History."
	}
	fmt.Print("Inserted Deal History ID's: ")
	fmt.Println(insertManyDealHistoryResult.InsertedIDs)

	return dealsArray, ""
}

func getDealsForStrategy(w http.ResponseWriter, r *http.Request) {
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
	deals := make([]models.Deal, 0)

	query := bson.M{}
	query["user_id"] = userId
	query["strategy_id"] = params["id"]
	status := r.URL.Query().Get("status")
	if status != "" {
		statusArr := strings.Split(status, ",")
		for _, sts := range statusArr {
			if sts != "running" && sts != "bought" && sts != "completed" && sts != "cancelled" {
				customError.S = "Invalid query paramter status. Allowed values: [running, bought, completed, cancelled]"
				helper.GetError(&customError, w)
				return
			}
		}
		query["status"] = bson.M{"$in": statusArr}
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

func getDealsForUser(w http.ResponseWriter, r *http.Request) {
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

	// create Deals array
	deals := make([]models.Deal, 0)

	query := bson.M{}
	query["user_id"] = userId

	status := r.URL.Query().Get("status")

	if status != "" {
		statusArr := strings.Split(status, ",")
		for _, sts := range statusArr {
			if sts != "running" && sts != "bought" && sts != "completed" && sts != "cancelled" {
				customError.S = "Invalid query paramter status. Allowed values: [running, bought, completed, cancelled]"
				helper.GetError(&customError, w)
				return
			}
		}
		query["status"] = bson.M{"$in": statusArr}
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

func modifyDeal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var incomingbody models.ManipulateDeal
	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&incomingbody)

	var customError models.ErrorString
	var params = mux.Vars(r)

	if params["id"] == "" {
		customError.S = "Deal Id cannot be empty"
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

	var deal models.Deal

	err = dealsCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId}).Decode(&deal)
	if err != nil {
		helper.GetError(err, w)
		return
	}
	// Cancel or Close a Deal
	if incomingbody.Status != "" {

		if deal.Status != "running" && deal.Status != "bought" {
			customError.S = fmt.Sprintf("Cannot cancel or close a %s deal", deal.Status)
			helper.GetError(&customError, w)
			return
		}
		if incomingbody.Status != "cancelled" && incomingbody.Status != "close_at_market_price" {
			customError.S = fmt.Sprintf("Invalid Status Value. Allowed Values: [cancelled, close_at_market_price]")
			helper.GetError(&customError, w)
			return
		}
		if incomingbody.Status == "cancelled" {
			updateResult := dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "cancelled", "deal_cancelled_by_user": true}}, options.FindOneAndUpdate().SetReturnDocument(1))
			err = updateResult.Decode(&deal)
		} else {
			updateResult := dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "completed", "deal_closed_at_market_price_by_user": true}}, options.FindOneAndUpdate().SetReturnDocument(1))
			err = updateResult.Decode(&deal)
		}
		if err != nil {
			helper.GetError(err, w)
			return
		}
		json.NewEncoder(w).Encode(models.CancelDealResponse{Cancelled: true})

	} else if incomingbody.TargetProfit != "" && incomingbody.StopLossPercent != "" && incomingbody.MaxSafetyTradeCount != 0 && incomingbody.MaxActiveSafetyTradeCount != 0 {
		// Edit a Deal
		if deal.Status != "running" && deal.Status != "bought" {
			customError.S = fmt.Sprintf("Cannot modify a %s deal", deal.Status)
			helper.GetError(&customError, w)
			return
		}
		updateResult := dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": bson.M{"max_active_safety_trade_count": incomingbody.MaxActiveSafetyTradeCount, "max_safety_trade_count": incomingbody.MaxSafetyTradeCount, "target_profit": incomingbody.TargetProfit, "stop_loss_percent": incomingbody.StopLossPercent, "deal_edited_by_user": true}}, options.FindOneAndUpdate().SetReturnDocument(1))
		err = updateResult.Decode(&deal)
		if err != nil {
			helper.GetError(err, w)
			return
		}
		json.NewEncoder(w).Encode(deal)
	} else {
		customError.S = fmt.Sprintf("Not enough parameters to perform required action on the Deal")
		helper.GetError(&customError, w)
		return
	}

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

	//r.HandleFunc("/StrategyService/api/v1/deals", createDeal).Methods("POST")
	r.HandleFunc("/StrategyService/api/v1/deals/{id}", getDealsForStrategy).Methods("GET")
	r.HandleFunc("/StrategyService/api/v1/deals", getDealsForUser).Methods("GET")
	r.HandleFunc("/StrategyService/api/v1/deals/{id}", modifyDeal).Methods("PUT")

	r.HandleFunc("/StrategyService/api/v1/accountinfo/{id}", getAccountInfo).Methods("GET")

	// set our port address
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
