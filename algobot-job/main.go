package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/shopspring/decimal"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/algobot-job/helper"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/algobot-job/models"
	orderpackage "github.com/vikjdk7/Algotrading-GoLang-Rest/algobot-job/orders"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	//Mongo Variables
	orderCollection    *mongo.Collection
	dealsCollection    *mongo.Collection
	strategyCollection *mongo.Collection

	// Base Order Variables
	base_order_size       float64
	target_profit_percent float64
	stop_loss_percent     float64
	asset                 string

	total_buying_price    float64
	total_buying_quantity int64
	target_profit_value   float64
	avg_buying_price      float64
	total_selling_price   float64

	//Safety Order Variables
	safety_order_size             float64
	max_safety_order_count        int64
	max_active_safety_order_count int64
	price_deviation               float64
	safety_order_step_scale       float64
	safety_order_volume_scale     float64

	previous_safety_order_buying_price float64
	previous_safety_order_volume       float64
	previous_safety_order_step_scale   float64
	previous_safety_order_deviation    float64

	//Exchange Variables
	alpaca_api_key    string
	alpaca_api_secret string
	alpaca_url        string

	//Deal Variables
	user_id       string
	exchange_id   string
	deal_id       string
	strategy_id   string
	strategy_name string

	clock  *alpaca.Clock
	orders []models.OrdersData
)

func init() {
	//Connect to mongoDB with helper class
	orderCollection, dealsCollection, strategyCollection = helper.ConnectDB()

	fmt.Println("--------------------Process Starting --------------------")
	/*
		os.Setenv("base_order_size", "15.0")
		os.Setenv("target_profit_percent", "1.0")
		os.Setenv("safety_order_size", "10.0")
		os.Setenv("max_safety_order_count", "4")
		os.Setenv("max_active_safety_order_count", "2")
		os.Setenv("price_deviation", "1.0")
		os.Setenv("safety_order_step_scale", "0.0")
		os.Setenv("safety_order_volume_scale", "1.5")
		os.Setenv("stop_loss_percent", "2.0")
		os.Setenv("asset", "ASMB")
		os.Setenv("alpaca_api_key", "PKX55YM3PYOL8T8LBNML")
		os.Setenv("alpaca_api_secret", "ftlFvZpoRC05VUfJdzPJAsyp4HXT6tKhvm8sqFGN")
		os.Setenv("alpaca_url", "https://paper-api.alpaca.markets")
		os.Setenv("user_id", "1234567890")
		os.Setenv("exchange_id", "2345678901")
		os.Setenv("deal_id", "60fa84ce53b41035b5f4e3fa")
		os.Setenv("strategy_id", "45678901")
	*/
	fmt.Println("--------------------Reading ENV Variables--------------------")
	base_order_size, _ = strconv.ParseFloat(os.Getenv("base_order_size"), 64)
	target_profit_percent, _ = strconv.ParseFloat(os.Getenv("target_profit_percent"), 64)
	safety_order_size, _ = strconv.ParseFloat(os.Getenv("safety_order_size"), 64)
	max_safety_order_count, _ = strconv.ParseInt(os.Getenv("max_safety_order_count"), 10, 64)
	max_active_safety_order_count, _ = strconv.ParseInt(os.Getenv("max_active_safety_order_count"), 10, 64)
	price_deviation, _ = strconv.ParseFloat(os.Getenv("price_deviation"), 64)
	safety_order_step_scale, _ = strconv.ParseFloat(os.Getenv("safety_order_step_scale"), 64)
	safety_order_volume_scale, _ = strconv.ParseFloat(os.Getenv("safety_order_volume_scale"), 64)
	stop_loss_percent, _ = strconv.ParseFloat(os.Getenv("stop_loss_percent"), 64)
	asset = os.Getenv("asset")

	previous_safety_order_step_scale = 0.0
	target_profit_value = 0.0

	//Alpaca Exchange Parameters
	alpaca_api_key = os.Getenv("alpaca_api_key")
	alpaca_api_secret = os.Getenv("alpaca_api_secret")
	alpaca_url = os.Getenv("alpaca_url")

	user_id = os.Getenv("user_id")
	exchange_id = os.Getenv("exchange_id")
	deal_id = os.Getenv("deal_id")
	strategy_id = os.Getenv("strategy_id")
	strategy_name = os.Getenv("strategy_name")
}

func main() {

	fmt.Println(fmt.Sprintf("Bot will run with Parameters: base_order_size: %v, target_profit_percent: %v, stop_loss_percent: %v, asset: %v, safety_order_size: %v, max_safety_order_count: %v, max_active_safety_order_count: %v, price_deviation: %v, safety_order_step_scale: %v, safety_order_volume_scale: %v, user_id: %v, exchange_id: %v, deal_id: %v, strategy_id: %v", base_order_size, target_profit_percent, stop_loss_percent, asset, safety_order_size, max_safety_order_count, max_active_safety_order_count, price_deviation, safety_order_step_scale, safety_order_volume_scale, user_id, exchange_id, deal_id, strategy_id))

	fmt.Println("--------------------Starting BOT--------------------")
	SetAlpacaParameters(alpaca_url, alpaca_api_key, alpaca_api_secret)

	alpacaClient := alpaca.NewClient(common.Credentials())

	// Check if the market is open now.
	clock, err := alpacaClient.GetClock()

	if err != nil {
		for {
			clock, err = alpacaClient.GetClock()
			if err == nil {
				break
			}
		}
	}

	//Waitgroup
	wg := &sync.WaitGroup{}

	if clock.IsOpen == false {
		fmt.Println("The market is closed.")
		t2 := clock.NextOpen
		t1 := time.Now()
		diff := t2.Sub(t1)
		fmt.Printf("Market will open after: %v", diff)
		fmt.Println(fmt.Sprintf("Sleeping for %v", diff))
		time.Sleep(diff)
		clock, _ = alpacaClient.GetClock()
	}

	fmt.Println("The market is open.")

	var order models.OrderMongo
	//Get Current AssetPrice
	initial_buying_price := GetCurrentAssetPrice(alpacaClient)
	//initial_buying_price := 3.90
	fmt.Println(fmt.Sprintf("Current Price of asset %v is %v", asset, initial_buying_price))

	//Calculate initial order quantity
	initial_order_quantity := int64(base_order_size)
	fmt.Println(fmt.Sprintf("initial_order_quantity is : %v", initial_order_quantity))

	total_buying_price = float64(initial_order_quantity) * initial_buying_price

	//Place a BUY "Market Order"
	fmt.Println("--------------------Step 1: Placing a BUY Market Order--------------------")

	orderPlaced, err := orderpackage.PlaceOrder(alpacaClient, asset, initial_order_quantity, "BUY", "market", 0.0, 0.0)
	if err != nil {
		panic(err)
	}

	for {
		orderDetails, _ := alpacaClient.GetOrder(orderPlaced.ID)
		fmt.Println(fmt.Sprintf("BUY Order Status: %v ", orderDetails.Status))
		if orderDetails.Status == "filled" {
			jsonOrder, _ := json.Marshal(orderDetails)
			_ = json.Unmarshal(jsonOrder, &order)
			order.Qty, _ = orderDetails.Qty.Float64()
			order.Notional, _ = orderDetails.Notional.Float64()
			order.FilledQty, _ = orderDetails.FilledQty.Float64()

			if orderDetails.LimitPrice != nil {
				order.LimitPrice, _ = orderDetails.LimitPrice.Float64()
			}
			if orderDetails.FilledAvgPrice != nil {
				order.FilledAvgPrice, _ = orderDetails.FilledAvgPrice.Float64()
			}
			if orderDetails.StopPrice != nil {
				order.StopPrice, _ = orderDetails.StopPrice.Float64()
			}
			order.UserId = user_id
			order.StrategyId = strategy_id
			order.ExchangeId = exchange_id
			order.DealId = deal_id
			order.StrategyName = strategy_name
			total_buying_quantity = orderDetails.FilledQty.IntPart()
			initial_buying_price, _ = (*orderDetails.FilledAvgPrice).Float64()
			avg_buying_price = initial_buying_price
			total_buying_price = initial_buying_price * float64(total_buying_quantity)
			for {
				_, err = orderCollection.InsertOne(context.TODO(), order)
				if err == nil {
					break
				}
			}
			break
		}

	}

	var orders_data models.OrdersData
	orders_data.OrderId = orderPlaced.ID
	orders_data.OrderType = "BUY"
	orders_data.OrderStatus = "filled"
	orders = append(orders, orders_data)
	fmt.Println(fmt.Sprintf("Current Orders ID list: %v ", orders))

	fmt.Println("--------------------Step 2: Placing a SELL Market Order--------------------")

	//Calculate target Profit value
	target_profit_value = (target_profit_percent * total_buying_price) / 100
	fmt.Println(fmt.Sprintf("Target profit value (in USD) to attain: %v ", target_profit_value))

	sell_limit_price := (total_buying_price + target_profit_value) / float64(total_buying_quantity)
	fmt.Println(fmt.Sprintf("SELL Limit Price: %v", sell_limit_price))
	total_selling_price = sell_limit_price * float64(total_buying_quantity)

	// Sell stop price = 2% less than sell limit price

	sell_stop_price := sell_limit_price * 0.98
	if current_asset_price := GetCurrentAssetPrice(alpacaClient); current_asset_price > sell_stop_price {
		sell_stop_price = current_asset_price
	}
	fmt.Println(fmt.Sprintf("Sell Stop Price(minimum of current asset price or 2 percent less than sell limit price): %v", sell_stop_price))

	sell_order_quantity := total_buying_quantity
	fmt.Println(fmt.Sprintf("Sell Order Quantity: %v", sell_order_quantity))

	//Place a SELL Order
	for {
		sellOrderPlaced, err := orderpackage.PlaceOrder(alpacaClient, asset, initial_order_quantity, "SELL", "stop_limit", sell_limit_price, sell_stop_price)
		if err == nil {
			//fmt.Println(fmt.Sprintf("SELL Order Placed: %v ", sellOrderPlaced))
			orders_data.OrderId = sellOrderPlaced.ID
			orders_data.OrderType = "SELL"
			orders_data.OrderStatus = "new"
			orders = append(orders, orders_data)
			jsonOrder, _ := json.Marshal(sellOrderPlaced)
			_ = json.Unmarshal(jsonOrder, &order)
			order.Qty, _ = sellOrderPlaced.Qty.Float64()
			order.Notional, _ = sellOrderPlaced.Notional.Float64()
			order.FilledQty, _ = sellOrderPlaced.FilledQty.Float64()

			if sellOrderPlaced.LimitPrice != nil {
				order.LimitPrice, _ = sellOrderPlaced.LimitPrice.Float64()
			}
			if sellOrderPlaced.FilledAvgPrice != nil {
				order.FilledAvgPrice, _ = sellOrderPlaced.FilledAvgPrice.Float64()
			}
			if sellOrderPlaced.StopPrice != nil {
				order.StopPrice, _ = sellOrderPlaced.StopPrice.Float64()
			}
			order.UserId = user_id
			order.StrategyId = strategy_id
			order.ExchangeId = exchange_id
			order.StrategyName = strategy_name
			order.DealId = deal_id
			for {
				_, err = orderCollection.InsertOne(context.TODO(), order)
				if err == nil {
					break
				}
			}
			break
		}
	}

	fmt.Println(fmt.Sprintf("Orders Array : %v ", orders))

	fmt.Println("--------------------Step 3: Placing Stop Limit BUY Safety Orders--------------------")
	previous_safety_order_buying_price = initial_buying_price
	//previous_safety_order_volume = safety_order_size
	previous_safety_order_deviation = price_deviation
	for i := int64(0); i < max_active_safety_order_count; i++ {
		safety_order_step := previous_safety_order_step_scale * safety_order_step_scale
		safety_order_deviation := previous_safety_order_deviation + safety_order_step
		safety_order_limit_price := previous_safety_order_buying_price * (1.0 - (safety_order_deviation / 100))
		safety_order_stop_price := safety_order_limit_price * 1.02
		if current_asset_price := GetCurrentAssetPrice(alpacaClient); current_asset_price < safety_order_stop_price {
			safety_order_stop_price = current_asset_price
		}

		var safety_order_quantity int64
		var safety_order_volume float64
		if i == 0 {
			safety_order_quantity = int64(safety_order_size)
			safety_order_volume = safety_order_size * safety_order_limit_price
		} else {
			safety_order_volume = previous_safety_order_volume * safety_order_volume_scale
			safety_order_quantity = int64(safety_order_volume / safety_order_limit_price)
		}

		//Place a BUY Safety Order
		fmt.Println(fmt.Sprintf("Safety Order %v. Safety order buying Price: %v, safety order step: %v, Safety Order total deviation: %v", i+1, previous_safety_order_buying_price, safety_order_step, safety_order_deviation))
		fmt.Printf("Placing Safety Order %v with limit price: %v, stop_price: %v, safety order volume: %v, safety_order_quantity: %v", i+1, safety_order_limit_price, safety_order_stop_price, safety_order_volume, safety_order_quantity)
		for {
			buySafetyOrderPlaced, err := orderpackage.PlaceOrder(alpacaClient, asset, safety_order_quantity, "BUY", "stop_limit", safety_order_limit_price, safety_order_stop_price)
			if err == nil {
				//fmt.Println(fmt.Sprintf("BUY Safety Order Placed: %v ", buySafetyOrderPlaced))
				orders_data.OrderId = buySafetyOrderPlaced.ID
				orders_data.OrderType = "SO"
				orders_data.OrderStatus = "new"
				orders = append(orders, orders_data)
				jsonOrder, _ := json.Marshal(buySafetyOrderPlaced)
				_ = json.Unmarshal(jsonOrder, &order)
				order.Qty, _ = buySafetyOrderPlaced.Qty.Float64()
				order.Notional, _ = buySafetyOrderPlaced.Notional.Float64()
				order.FilledQty, _ = buySafetyOrderPlaced.FilledQty.Float64()

				if buySafetyOrderPlaced.LimitPrice != nil {
					order.LimitPrice, _ = buySafetyOrderPlaced.LimitPrice.Float64()
				}
				if buySafetyOrderPlaced.FilledAvgPrice != nil {
					order.FilledAvgPrice, _ = buySafetyOrderPlaced.FilledAvgPrice.Float64()
				}
				if buySafetyOrderPlaced.StopPrice != nil {
					order.StopPrice, _ = buySafetyOrderPlaced.StopPrice.Float64()
				}
				order.UserId = user_id
				order.StrategyId = strategy_id
				order.ExchangeId = exchange_id
				order.StrategyName = strategy_name
				order.DealId = deal_id
				for {
					_, err = orderCollection.InsertOne(context.TODO(), order)
					if err == nil {
						break
					}
				}
				break
			}
		}

		previous_safety_order_buying_price = safety_order_limit_price
		previous_safety_order_volume = safety_order_volume
		previous_safety_order_deviation = safety_order_deviation
		if i == 0 {
			previous_safety_order_step_scale = price_deviation
		} else {
			previous_safety_order_step_scale = safety_order_step
		}
	}
	fmt.Println(fmt.Sprintf("Orders Array : %v ", orders))

	updateDeal := bson.M{}
	next_safety_order_limit_price := previous_safety_order_buying_price * (1.0 - ((previous_safety_order_deviation + (previous_safety_order_step_scale * safety_order_step_scale)) / 100))

	updateDeal["active_safety_order_count"] = max_active_safety_order_count
	updateDeal["total_order_quantity"] = total_buying_quantity
	updateDeal["total_buying_price"] = total_buying_price
	updateDeal["total_sell_price"] = total_selling_price
	updateDeal["next_safety_order_limit_price"] = next_safety_order_limit_price
	updateDeal["status"] = "bought"
	updateDeal["avg_buying_price"] = avg_buying_price
	fmt.Println("Updating Deal")
	id, _ := primitive.ObjectIDFromHex(deal_id)
	for {
		result := dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": updateDeal})
		if result.Err() == nil {
			break
		}
	}

	dealCancelled := make(chan bool)
	dealClosedAtMarketPrice := make(chan bool)
	dealEdited := make(chan bool)
	dealBuyMore := make(chan bool)
	dealCompleted := make(chan bool)
	dealEditCompleted := make(chan bool)
	go CheckDealStatus(deal_id, dealCancelled, dealClosedAtMarketPrice, dealEdited, dealBuyMore, dealEditCompleted)
	go CheckOrderStatus(alpacaClient, dealCompleted)
F:
	for range time.Tick(time.Second * 2) {
	S:
		select {
		case <-dealCancelled:
			fmt.Println("Deal Cancelled manually by user")
			CancelAllNonFilledOrders(alpacaClient)
			break F

		case <-dealClosedAtMarketPrice:
			fmt.Println("Deal closed at market Price by user")
			CloseAtMarketPrice(alpacaClient)
			break F

		case <-dealCompleted:
			fmt.Println("Target Profit Reached. Cancelling unfilled orders.")
			CalculateProfitAndCancelUnfilledOrders(alpacaClient)
			break F

		case <-dealBuyMore:
			fmt.Println("More Stocks bought by the user")
			HandleBuyMoreStocks(alpacaClient)
			break S

		case <-dealEdited:
			fmt.Println("Deal Edited by User")
			HandleDealEdit(alpacaClient, dealEditCompleted)
			break S

		default:

		}
	}

	wg.Wait()
}

func SetAlpacaParameters(baseUrl string, api_key string, api_secret string) {
	os.Setenv(common.EnvApiKeyID, api_key)
	os.Setenv(common.EnvApiSecretKey, api_secret)
	alpaca.SetBaseUrl(baseUrl)
}

func GetCurrentAssetPrice(alpacaClient *alpaca.Client) float64 {
	latestQuote, err := alpacaClient.GetLatestQuote(asset)
	if err != nil || latestQuote.AskPrice == 0.0 {
		for {
			latestQuote, err = alpacaClient.GetLatestQuote(asset)
			if err == nil && latestQuote.AskPrice != 0.0 {
				break
			}
		}
	}
	return latestQuote.AskPrice
}

func CheckOrderStatus(alpacaClient *alpaca.Client, dealCompleted chan bool) {
	fmt.Println("Checking Order Status Every Second")

	var dealCompletedVar bool
	for range time.Tick(time.Second * 2) {

		for _, o := range orders {
			o.Mu.Lock()
			if o.OrderType == "SELL" {
				sellOrderId := orders[1].OrderId
				sellOrderDetails, err := alpacaClient.GetOrder(sellOrderId)
				if err != nil && sellOrderDetails.Status == "filled" {
					orders[1].OrderStatus = "filled"
					dealCompletedVar = true
					dealCompleted <- true
					break
				}
			} else if o.OrderType == "SO" && o.OrderStatus != "filled" {
				fmt.Println(fmt.Sprintf("Checking Order Status for Order Id: %v", o.OrderId))
				orderDetails, err := alpacaClient.GetOrder(o.OrderId)
				if err != nil && orderDetails.Status == "filled" {
					var order models.OrderMongo
					fmt.Println("Safety Order %v is filled", o.OrderId)
					fmt.Println("Step 1: Updating SELL Order")
					so_quantity := orderDetails.FilledQty.IntPart()
					so_price, _ := (*orderDetails.FilledAvgPrice).Float64()
					total_so_price := so_price * float64(so_quantity)
					avg_buying_price = (total_so_price + total_buying_price) / float64(total_buying_quantity+so_quantity)
					total_buying_price = total_buying_price + total_so_price
					total_buying_quantity = total_buying_quantity + so_quantity
					new_sell_order_limit_price := (total_buying_price + target_profit_value) / float64(total_buying_quantity)
					new_sell_order_stop_price := new_sell_order_limit_price * 0.98
					if current_asset_price := GetCurrentAssetPrice(alpacaClient); current_asset_price > new_sell_order_stop_price {
						new_sell_order_stop_price = current_asset_price
					}
					total_selling_price = new_sell_order_limit_price * float64(total_buying_quantity)

					fmt.Println(fmt.Sprintf("Cancelling Old SELL Order: %v ", orders[1].OrderId))
					CancelOrder(orders[1].OrderId, alpacaClient)
					fmt.Println(fmt.Sprintf("Placing new SELL Order with Qty: %v, Limit Price: %v, Stop Price: %v", total_buying_quantity, new_sell_order_limit_price, new_sell_order_stop_price))

					//Place a SELL Order
					sellOrderPlaced, err := orderpackage.PlaceOrder(alpacaClient, asset, total_buying_quantity, "SELL", "stop_limit", new_sell_order_limit_price, new_sell_order_stop_price)
					if err == nil {
						//fmt.Println(fmt.Sprintf("SELL Order Placed: %v ", sellOrderPlaced))
						orders[1].OrderId = sellOrderPlaced.ID

						jsonOrder, _ := json.Marshal(sellOrderPlaced)
						_ = json.Unmarshal(jsonOrder, &order)
						order.Qty, _ = sellOrderPlaced.Qty.Float64()
						order.Notional, _ = sellOrderPlaced.Notional.Float64()
						order.FilledQty, _ = sellOrderPlaced.FilledQty.Float64()

						if sellOrderPlaced.LimitPrice != nil {
							order.LimitPrice, _ = sellOrderPlaced.LimitPrice.Float64()
						}
						if sellOrderPlaced.FilledAvgPrice != nil {
							order.FilledAvgPrice, _ = sellOrderPlaced.FilledAvgPrice.Float64()
						}
						if sellOrderPlaced.StopPrice != nil {
							order.StopPrice, _ = sellOrderPlaced.StopPrice.Float64()
						}
						order.UserId = user_id
						order.StrategyId = strategy_id
						order.ExchangeId = exchange_id
						order.DealId = deal_id
						order.StrategyName = strategy_name
						for {
							_, err = orderCollection.InsertOne(context.TODO(), order)
							if err == nil {
								break
							}
						}
					}
					id, _ := primitive.ObjectIDFromHex(deal_id)
					_ = dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": bson.M{"avg_buying_price": avg_buying_price, "total_buying_price": total_buying_price, "total_sell_price": total_selling_price, "total_order_quantity": total_buying_quantity}})

					fmt.Println("Step 2: Checking if more Safety Orders can be placed")
					if int64(len(orders)-2) < max_safety_order_count {
						fmt.Println("Creating new Safety Order")
						safety_order_step := previous_safety_order_step_scale * safety_order_step_scale
						safety_order_deviation := previous_safety_order_deviation + safety_order_step
						safety_order_limit_price := previous_safety_order_buying_price * (1.0 - (safety_order_deviation / 100))
						safety_order_stop_price := safety_order_limit_price * 1.02
						if current_asset_price := GetCurrentAssetPrice(alpacaClient); current_asset_price < safety_order_stop_price {
							safety_order_stop_price = current_asset_price
						}

						safety_order_volume := previous_safety_order_volume * safety_order_volume_scale

						safety_order_quantity := int64(safety_order_volume / safety_order_limit_price)
						fmt.Printf("Placing New Safety Order with limit price: %v, stop_price: %v, safety order volume: %v, safety_order_quantity: %v", safety_order_limit_price, safety_order_stop_price, safety_order_volume, safety_order_quantity)
						buySafetyOrderPlaced, _ := orderpackage.PlaceOrder(alpacaClient, asset, safety_order_quantity, "BUY", "stop_limit", safety_order_limit_price, safety_order_stop_price)
						var orders_data models.OrdersData
						orders_data.OrderId = buySafetyOrderPlaced.ID
						orders_data.OrderType = "SO"
						orders_data.OrderStatus = "new"
						orders = append(orders, orders_data)
						jsonOrder, _ := json.Marshal(buySafetyOrderPlaced)
						_ = json.Unmarshal(jsonOrder, &order)
						order.Qty, _ = buySafetyOrderPlaced.Qty.Float64()
						order.Notional, _ = buySafetyOrderPlaced.Notional.Float64()
						order.FilledQty, _ = buySafetyOrderPlaced.FilledQty.Float64()

						if buySafetyOrderPlaced.LimitPrice != nil {
							order.LimitPrice, _ = buySafetyOrderPlaced.LimitPrice.Float64()
						}
						if buySafetyOrderPlaced.FilledAvgPrice != nil {
							order.FilledAvgPrice, _ = buySafetyOrderPlaced.FilledAvgPrice.Float64()
						}
						if buySafetyOrderPlaced.StopPrice != nil {
							order.StopPrice, _ = buySafetyOrderPlaced.StopPrice.Float64()
						}
						order.UserId = user_id
						order.StrategyId = strategy_id
						order.ExchangeId = exchange_id
						order.DealId = deal_id
						order.StrategyName = strategy_name
						for {
							_, err = orderCollection.InsertOne(context.TODO(), order)
							if err == nil {
								break
							}
						}
						previous_safety_order_buying_price = safety_order_limit_price
						previous_safety_order_volume = safety_order_volume
						previous_safety_order_deviation = safety_order_deviation
						previous_safety_order_step_scale = safety_order_step
					} else if orders[len(orders)-1].OrderStatus == "filled" {
						fmt.Println("All Safety Orders are filled. SELLing at a LOSS")
						CloseAtMarketPrice(alpacaClient)
						dealCompletedVar = true
					}
					o.OrderStatus = "filled"
				}
			}
			o.Mu.Unlock()
		}
		if dealCompletedVar == true {
			break
		}
	}
}

func CancelOrder(orderid string, alpacaClient *alpaca.Client) {
	err := alpacaClient.CancelOrder(orderid)
	if err != nil {
		orderDetails, _ := alpacaClient.GetOrder(orderid)
		if orderDetails.Status != "canceled" {
			// RETRY Cancelling
			for {
				err = alpacaClient.CancelOrder(orderid)
				if err == nil {
					break
				}
			}
		}
	}
	orderDetails, err := alpacaClient.GetOrder(orderid)
	if err != nil {
		for {
			orderDetails, err = alpacaClient.GetOrder(orderid)
			if err == nil {
				break
			}
		}
	}
	if orderDetails.Status == "canceled" {
		var cancelledOrder models.OrderMongo
		jsonOrder, _ := json.Marshal(orderDetails)
		_ = json.Unmarshal(jsonOrder, &cancelledOrder)
		cancelledOrder.Qty, _ = orderDetails.Qty.Float64()
		cancelledOrder.Notional, _ = orderDetails.Notional.Float64()
		cancelledOrder.FilledQty, _ = orderDetails.FilledQty.Float64()

		if orderDetails.LimitPrice != nil {
			cancelledOrder.LimitPrice, _ = orderDetails.LimitPrice.Float64()
		}
		if orderDetails.FilledAvgPrice != nil {
			cancelledOrder.FilledAvgPrice, _ = orderDetails.FilledAvgPrice.Float64()
		}
		if orderDetails.StopPrice != nil {
			cancelledOrder.StopPrice, _ = orderDetails.StopPrice.Float64()
		}
		cancelledOrder.UserId = user_id
		cancelledOrder.StrategyId = strategy_id
		cancelledOrder.ExchangeId = exchange_id
		cancelledOrder.DealId = deal_id
		cancelledOrder.StrategyName = strategy_name
		_ = orderCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": orderid}, bson.M{"$set": cancelledOrder}, options.FindOneAndUpdate().SetReturnDocument(1))
	}
}

func CheckDealStatus(deal_id string, dealCancelled, dealClosedAtMarketPrice, dealEdited, dealBuyMore, dealEditCompleted chan bool) {
	var deal models.Deal
	id, _ := primitive.ObjectIDFromHex(deal_id)

	fmt.Println("Continuously checking deal status every second from Database")
	for range time.Tick(time.Second * 2) {
		err := dealsCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&deal)
		if err != nil {
			fmt.Println(err)
			//RETRY
			for {
				err = dealsCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&deal)
				if err == nil {
					break
				}
			}
		}
		if deal.DealCancelledByUser == true {
			dealCancelled <- true
			break
		} else if deal.DealClosedAtMarketPriceByUser == true {
			dealClosedAtMarketPrice <- true
			break
		} else if deal.ManualOrderPlacedByUser == true {
			dealBuyMore <- true
		} else if deal.DealEditedByUser == true {
			fmt.Println("Deal Edit Signal Detected")
			dealEdited <- true
			<-dealEditCompleted
		}
	}
}

func HandleBuyMoreStocks(alpacaClient *alpaca.Client) {

	cur, err := orderCollection.Find(context.TODO(), bson.M{"deal_id": deal_id, "side": "buy"})
	if err != nil {
		for {
			cur, err = orderCollection.Find(context.TODO(), bson.M{"deal_id": deal_id, "side": "buy"})
			if err == nil {
				break
			}
		}
	}
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var order models.OrderMongo
		_ = cur.Decode(&order)
		var found bool
		for _, o := range orders {
			if order.ID == o.OrderId {
				found = true
				break
			}
		}
		if found == false {
			fmt.Println(fmt.Sprintf("New Order %v was manually placed by user", order.ID))
			orderDetails, _ := alpacaClient.GetOrder(order.ID)
			if orderDetails.Status != "filled" {
				fmt.Println("Waiting for Order to be filled")
				for {
					orderDetails, _ := alpacaClient.GetOrder(order.ID)
					if orderDetails.Status == "filled" {
						break
					}
				}
			}

			fmt.Println("Updating Total Order Quantity")
			new_order_qty := orderDetails.FilledQty.IntPart()
			new_order_price, _ := (*orderDetails.FilledAvgPrice).Float64()
			//new_order_price, _ := (*orderDetails.FilledAvgPrice).Float64() * float64(new_order_qty)
			avg_buying_price = ((new_order_price * float64(new_order_qty)) + total_buying_price) / float64(total_buying_quantity+new_order_qty)
			total_buying_price = total_buying_price + (new_order_price * float64(new_order_qty))
			total_buying_quantity = total_buying_quantity + new_order_qty
			target_profit_value = (total_buying_price * target_profit_percent) / 100
			new_sell_order_limit_price := (total_buying_price + target_profit_value) / float64(total_buying_quantity)
			total_selling_price = new_sell_order_limit_price * float64(total_buying_quantity)
			new_sell_order_stop_price := new_sell_order_limit_price * 0.98
			if current_asset_price := GetCurrentAssetPrice(alpacaClient); current_asset_price > new_sell_order_stop_price {
				new_sell_order_stop_price = current_asset_price
			}
			fmt.Println(fmt.Sprintf("Cancelling Old SELL Order: %v ", orders[1].OrderId))
			CancelOrder(orders[1].OrderId, alpacaClient)
			fmt.Println(fmt.Sprintf("Placing new SELL Order with Qty: %v, Limit Price: %v, Stop Price: %v", total_buying_quantity, new_sell_order_limit_price, new_sell_order_stop_price))

			//Place a SELL Order
			sellOrderPlaced, err := orderpackage.PlaceOrder(alpacaClient, asset, total_buying_quantity, "SELL", "stop_limit", new_sell_order_limit_price, new_sell_order_stop_price)
			if err == nil {
				//fmt.Println(fmt.Sprintf("SELL Order Placed: %v ", sellOrderPlaced))
				orders[1].OrderId = sellOrderPlaced.ID

				jsonOrder, _ := json.Marshal(sellOrderPlaced)
				_ = json.Unmarshal(jsonOrder, &order)
				order.Qty, _ = sellOrderPlaced.Qty.Float64()
				order.Notional, _ = sellOrderPlaced.Notional.Float64()
				order.FilledQty, _ = sellOrderPlaced.FilledQty.Float64()

				if sellOrderPlaced.LimitPrice != nil {
					order.LimitPrice, _ = sellOrderPlaced.LimitPrice.Float64()
				}
				if sellOrderPlaced.FilledAvgPrice != nil {
					order.FilledAvgPrice, _ = sellOrderPlaced.FilledAvgPrice.Float64()
				}
				if sellOrderPlaced.StopPrice != nil {
					order.StopPrice, _ = sellOrderPlaced.StopPrice.Float64()
				}
				order.UserId = user_id
				order.StrategyId = strategy_id
				order.ExchangeId = exchange_id
				order.DealId = deal_id
				order.StrategyName = strategy_name
				for {
					_, err = orderCollection.InsertOne(context.TODO(), order)
					if err == nil {
						break
					}
				}
			}

			fmt.Println("Updating Parameters for Safety Orders")

			if previous_safety_order_buying_price > avg_buying_price {
				previous_safety_order_buying_price = avg_buying_price
				for _, v := range orders {
					if v.OrderType == "SO" && v.OrderStatus != "filled" {
						fmt.Println(fmt.Sprintf("Cancelling Safety Order: %v", v.OrderId))
						CancelOrder(v.OrderId, alpacaClient)
						fmt.Println("Placing new Safety Order Instead")
						safety_order_step := previous_safety_order_step_scale * safety_order_step_scale
						safety_order_deviation := previous_safety_order_deviation + safety_order_step
						safety_order_limit_price := previous_safety_order_buying_price * (1.0 - (safety_order_deviation / 100))
						safety_order_stop_price := safety_order_limit_price * 1.02
						if current_asset_price := GetCurrentAssetPrice(alpacaClient); current_asset_price < safety_order_stop_price {
							safety_order_stop_price = current_asset_price
						}

						safety_order_volume := previous_safety_order_volume * safety_order_volume_scale

						safety_order_quantity := int64(safety_order_volume / safety_order_limit_price)
						fmt.Printf("Placing Safety Order with limit price: %v, stop_price: %v, safety order volume: %v, safety_order_quantity: %v", safety_order_limit_price, safety_order_stop_price, safety_order_volume, safety_order_quantity)
						buySafetyOrderPlaced, _ := orderpackage.PlaceOrder(alpacaClient, asset, safety_order_quantity, "BUY", "stop_limit", safety_order_limit_price, safety_order_stop_price)
						v.OrderId = buySafetyOrderPlaced.ID
						v.OrderStatus = "new"
						jsonOrder, _ := json.Marshal(buySafetyOrderPlaced)
						_ = json.Unmarshal(jsonOrder, &order)
						order.Qty, _ = buySafetyOrderPlaced.Qty.Float64()
						order.Notional, _ = buySafetyOrderPlaced.Notional.Float64()
						order.FilledQty, _ = buySafetyOrderPlaced.FilledQty.Float64()

						if buySafetyOrderPlaced.LimitPrice != nil {
							order.LimitPrice, _ = buySafetyOrderPlaced.LimitPrice.Float64()
						}
						if buySafetyOrderPlaced.FilledAvgPrice != nil {
							order.FilledAvgPrice, _ = buySafetyOrderPlaced.FilledAvgPrice.Float64()
						}
						if buySafetyOrderPlaced.StopPrice != nil {
							order.StopPrice, _ = buySafetyOrderPlaced.StopPrice.Float64()
						}
						order.UserId = user_id
						order.StrategyId = strategy_id
						order.ExchangeId = exchange_id
						order.DealId = deal_id
						order.StrategyName = strategy_name
						for {
							_, err = orderCollection.InsertOne(context.TODO(), order)
							if err == nil {
								break
							}
						}
						previous_safety_order_buying_price = safety_order_limit_price
						previous_safety_order_volume = safety_order_volume
						previous_safety_order_deviation = safety_order_deviation
						previous_safety_order_step_scale = safety_order_step
					}
				}
			}

			var orders_data models.OrdersData
			orders_data.OrderId = order.ID
			orders_data.OrderType = "BUY"
			orders_data.OrderStatus = "filled"
			id, _ := primitive.ObjectIDFromHex(deal_id)
			_ = dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": bson.M{"avg_buying_price": avg_buying_price, "manual_order_placed_by_user": false, "total_buying_price": total_buying_price, "total_sell_price": total_selling_price, "total_order_quantity": total_buying_quantity}})

		}
	}
}

func CalculateProfitAndCancelUnfilledOrders(alpacaClient *alpaca.Client) {
	for _, v := range orders {
		if v.OrderStatus != "filled" && v.OrderType == "SO" {
			for {
				fmt.Printf("Cancelling Safety Order %s", v.OrderId)
				err := alpacaClient.CancelOrder(v.OrderId)
				if err != nil {
					orderDetails, _ := alpacaClient.GetOrder(v.OrderId)
					if orderDetails.Status != "canceled" {
						// RETRY Cancelling
						for {
							err = alpacaClient.CancelOrder(v.OrderId)
							if err == nil {
								break
							}
						}
					}
				}
				orderDetails, err := alpacaClient.GetOrder(v.OrderId)
				if err != nil {
					for {
						orderDetails, err = alpacaClient.GetOrder(v.OrderId)
						if err == nil {
							break
						}
					}
				}
				if orderDetails.Status == "canceled" {
					var cancelledOrder models.OrderMongo
					jsonOrder, _ := json.Marshal(orderDetails)
					_ = json.Unmarshal(jsonOrder, &cancelledOrder)
					cancelledOrder.Qty, _ = orderDetails.Qty.Float64()
					cancelledOrder.Notional, _ = orderDetails.Notional.Float64()
					cancelledOrder.FilledQty, _ = orderDetails.FilledQty.Float64()

					if orderDetails.LimitPrice != nil {
						cancelledOrder.LimitPrice, _ = orderDetails.LimitPrice.Float64()
					}
					if orderDetails.FilledAvgPrice != nil {
						cancelledOrder.FilledAvgPrice, _ = orderDetails.FilledAvgPrice.Float64()
					}
					if orderDetails.StopPrice != nil {
						cancelledOrder.StopPrice, _ = orderDetails.StopPrice.Float64()
					}
					cancelledOrder.UserId = user_id
					cancelledOrder.StrategyId = strategy_id
					cancelledOrder.ExchangeId = exchange_id
					cancelledOrder.DealId = deal_id
					cancelledOrder.StrategyName = strategy_name
					_ = orderCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": v.OrderId}, bson.M{"$set": cancelledOrder}, options.FindOneAndUpdate().SetReturnDocument(1))
					break
				}
			}
		}
	}
	id, _ := primitive.ObjectIDFromHex(deal_id)
	update := bson.M{}
	update["status"] = "completed"

	sellOrder, err := alpacaClient.GetOrder(orders[1].OrderId)
	if err != nil {
		for {
			sellOrder, err = alpacaClient.GetOrder(orders[1].OrderId)
			if err == nil {
				break
			}
		}
	}
	total_selling_quantity := sellOrder.FilledQty.IntPart()
	selling_price, _ := (*sellOrder.FilledAvgPrice).Float64()
	total_selling_price = selling_price * float64(total_selling_quantity)
	profit_value := total_selling_price - total_buying_price
	profit_percentage := (profit_value / total_buying_price) * 100

	update["profit_percentage"] = fmt.Sprintf("%.5f", profit_percentage)
	update["profit_value"] = profit_value
	update["total_sell_price"] = total_selling_price
	update["active_safety_order_count"] = 0
	update["closed_at"] = time.Now().Format(time.RFC3339)
	for {
		result := dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": update})
		if result.Err() == nil {
			break
		}
	}
	var strategy models.Strategy
	strId, _ := primitive.ObjectIDFromHex(strategy_id)
	_ = strategyCollection.FindOne(context.TODO(), bson.M{"_id": strId}).Decode(&strategy)
	updateStr := bson.M{}
	updateStr["active_deals"] = strategy.ActiveDeals - 1
	if strategy.ActiveDeals-1 == 0 {
		updateStr["status"] = "completed"
	}
	_ = strategyCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": strId}, bson.M{"$set": updateStr})
}

func CancelAllNonFilledOrders(alpacaClient *alpaca.Client) {
	for _, v := range orders {
		if v.OrderStatus != "filled" {
			for {
				fmt.Printf("Cancel Order %s", v.OrderId)
				err := alpacaClient.CancelOrder(v.OrderId)
				if err != nil {
					orderDetails, _ := alpacaClient.GetOrder(v.OrderId)
					if orderDetails.Status != "canceled" {
						// RETRY Cancelling
						for {
							err = alpacaClient.CancelOrder(v.OrderId)
							if err == nil {
								break
							}
						}
					}
				}
				orderDetails, err := alpacaClient.GetOrder(v.OrderId)
				if err != nil {
					for {
						orderDetails, err = alpacaClient.GetOrder(v.OrderId)
						if err == nil {
							break
						}
					}
				}
				if orderDetails.Status == "canceled" {
					var cancelledOrder models.OrderMongo
					jsonOrder, _ := json.Marshal(orderDetails)
					_ = json.Unmarshal(jsonOrder, &cancelledOrder)
					cancelledOrder.Qty, _ = orderDetails.Qty.Float64()
					cancelledOrder.Notional, _ = orderDetails.Notional.Float64()
					cancelledOrder.FilledQty, _ = orderDetails.FilledQty.Float64()

					if orderDetails.LimitPrice != nil {
						cancelledOrder.LimitPrice, _ = orderDetails.LimitPrice.Float64()
					}
					if orderDetails.FilledAvgPrice != nil {
						cancelledOrder.FilledAvgPrice, _ = orderDetails.FilledAvgPrice.Float64()
					}
					if orderDetails.StopPrice != nil {
						cancelledOrder.StopPrice, _ = orderDetails.StopPrice.Float64()
					}
					cancelledOrder.UserId = user_id
					cancelledOrder.StrategyId = strategy_id
					cancelledOrder.ExchangeId = exchange_id
					cancelledOrder.DealId = deal_id
					cancelledOrder.StrategyName = strategy_name
					_ = orderCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": v.OrderId}, bson.M{"$set": cancelledOrder}, options.FindOneAndUpdate().SetReturnDocument(1))
					break
				}
			}
		}
	}
	id, _ := primitive.ObjectIDFromHex(deal_id)
	update := bson.M{}
	update["status"] = "cancelled"
	update["profit_percentage"] = "-100"
	update["closed_at"] = time.Now().Format(time.RFC3339)
	update["profit_value"] = (-1.0) * total_buying_price
	update["active_safety_order_count"] = 0
	for {
		result := dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": update})
		if result.Err() == nil {
			break
		}
	}
	var strategy models.Strategy
	strId, _ := primitive.ObjectIDFromHex(strategy_id)
	_ = strategyCollection.FindOne(context.TODO(), bson.M{"_id": strId}).Decode(&strategy)
	updateStr := bson.M{}
	updateStr["active_deals"] = strategy.ActiveDeals - 1
	if strategy.ActiveDeals-1 == 0 {
		updateStr["status"] = "completed"
	}
	_ = strategyCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": strId}, bson.M{"$set": updateStr})

}

func CloseAtMarketPrice(alpacaClient *alpaca.Client) {
	// Check if the market is open now.

	clock, err := alpacaClient.GetClock()
	if err != nil {
		for {
			clock, err = alpacaClient.GetClock()
			if err == nil {
				break
			}
		}
	}

	if clock.IsOpen == false {
		fmt.Println("The market is closed.")
		t2 := clock.NextOpen
		t1 := time.Now()
		diff := t2.Sub(t1)
		fmt.Printf("Market will open after: %v", diff)
		fmt.Println(fmt.Sprintf("Sleeping for %v", diff))
		time.Sleep(diff)
	}

	var sellOrder models.OrderMongo

	for _, v := range orders {
		if v.OrderStatus != "filled" {
			if v.OrderType == "SELL" {
				orderplaced, err := orderpackage.PlaceOrder(alpacaClient, asset, total_buying_quantity, "SELL", "market", 0.0, 0.0)
				if err == nil {
					for {
						orderDetails, _ := alpacaClient.GetOrder(orderplaced.ID)
						if orderDetails.Status == "filled" {
							jsonOrder, _ := json.Marshal(orderDetails)
							_ = json.Unmarshal(jsonOrder, &sellOrder)
							sellOrder.Qty, _ = orderDetails.Qty.Float64()
							sellOrder.Notional, _ = orderDetails.Notional.Float64()
							sellOrder.FilledQty, _ = orderDetails.FilledQty.Float64()

							if orderDetails.LimitPrice != nil {
								sellOrder.LimitPrice, _ = orderDetails.LimitPrice.Float64()
							}
							if orderDetails.FilledAvgPrice != nil {
								sellOrder.FilledAvgPrice, _ = orderDetails.FilledAvgPrice.Float64()
							}
							if orderDetails.StopPrice != nil {
								sellOrder.StopPrice, _ = orderDetails.StopPrice.Float64()
							}
							break
						}
					}

					sellOrder.UserId = user_id
					sellOrder.StrategyId = strategy_id
					sellOrder.ExchangeId = exchange_id
					sellOrder.DealId = deal_id
					sellOrder.StrategyName = strategy_name
					for {
						_, err = orderCollection.InsertOne(context.TODO(), sellOrder)
						if err == nil {
							break
						}
					}
				}

			}
			for {
				fmt.Printf("Cancel Order %s", v.OrderId)
				_ = alpacaClient.CancelOrder(v.OrderId)
				orderDetails, _ := alpacaClient.GetOrder(v.OrderId)
				var cancelledOrder models.OrderMongo
				jsonOrder, _ := json.Marshal(orderDetails)
				_ = json.Unmarshal(jsonOrder, &cancelledOrder)
				cancelledOrder.Qty, _ = orderDetails.Qty.Float64()
				cancelledOrder.Notional, _ = orderDetails.Notional.Float64()
				cancelledOrder.FilledQty, _ = orderDetails.FilledQty.Float64()

				if orderDetails.LimitPrice != nil {
					cancelledOrder.LimitPrice, _ = orderDetails.LimitPrice.Float64()
				}
				if orderDetails.FilledAvgPrice != nil {
					cancelledOrder.FilledAvgPrice, _ = orderDetails.FilledAvgPrice.Float64()
				}
				if orderDetails.StopPrice != nil {
					cancelledOrder.StopPrice, _ = orderDetails.StopPrice.Float64()
				}
				cancelledOrder.UserId = user_id
				cancelledOrder.StrategyId = strategy_id
				cancelledOrder.ExchangeId = exchange_id
				cancelledOrder.DealId = deal_id
				cancelledOrder.StrategyName = strategy_name
				if orderDetails.Status == "canceled" {
					_ = orderCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": v.OrderId}, bson.M{"$set": cancelledOrder}, options.FindOneAndUpdate().SetReturnDocument(1))
					break
				}
			}
		}

	}
	total_selling_quantity := int(sellOrder.FilledQty)
	selling_price := sellOrder.FilledAvgPrice
	total_selling_price := selling_price * float64(total_selling_quantity)
	profit_value := total_selling_price - total_buying_price
	profit_percentage := (profit_value / total_buying_price) * 100

	id, _ := primitive.ObjectIDFromHex(deal_id)
	update := bson.M{}
	update["status"] = "completed"
	update["profit_percentage"] = fmt.Sprintf("%.5f", profit_percentage)
	update["profit_value"] = profit_value
	update["total_sell_price"] = total_selling_price
	update["active_safety_order_count"] = 0
	update["closed_at"] = time.Now().Format(time.RFC3339)

	for {
		result := dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": update})
		if result.Err() == nil {
			break
		}
	}
	var strategy models.Strategy
	strId, _ := primitive.ObjectIDFromHex(strategy_id)
	_ = strategyCollection.FindOne(context.TODO(), bson.M{"_id": strId}).Decode(&strategy)
	updateStr := bson.M{}
	updateStr["active_deals"] = strategy.ActiveDeals - 1
	if strategy.ActiveDeals-1 == 0 {
		updateStr["status"] = "completed"
	}
	_ = strategyCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": strId}, bson.M{"$set": updateStr})

}

func HandleDealEdit(alpacaClient *alpaca.Client, dealEditCompleted chan bool) {
	// Check if the market is open now.

	clock, err := alpacaClient.GetClock()
	if err != nil {
		for {
			clock, err = alpacaClient.GetClock()
			if err == nil {
				break
			}
		}
	}
	if clock.IsOpen == false {
		fmt.Println("The market is closed.")
		t2 := clock.NextOpen
		t1 := time.Now()
		diff := t2.Sub(t1)
		fmt.Printf("Market will open after: %v", diff)
		fmt.Println(fmt.Sprintf("Sleeping for %v", diff))
		time.Sleep(diff)
	}

	var deal models.Deal

	id, _ := primitive.ObjectIDFromHex(deal_id)
	err = dealsCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&deal)
	if err != nil {
		fmt.Println(err)
		//RETRY
		for {
			err = dealsCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&deal)
			if err == nil {
				break
			}
		}
	}
	updateDeal := bson.M{}

	if targetProfit, _ := strconv.ParseFloat(deal.TargetProfit, 64); targetProfit != target_profit_percent {
		target_profit_percent = targetProfit
		target_profit_value = (target_profit_percent * total_buying_price) / 100
		fmt.Println(fmt.Sprintf("New Target profit value (in USD) to attain: %v ", target_profit_value))
		sell_limit_price := (total_buying_price + target_profit_value) / float64(total_buying_quantity)
		fmt.Println(fmt.Sprintf("New SELL Limit Price: %v", sell_limit_price))
		total_selling_price = sell_limit_price * float64(total_buying_quantity)

		// Sell stop price = 2% less than sell limit price
		sell_stop_price := sell_limit_price * 0.98
		current_asset_price := GetCurrentAssetPrice(alpacaClient)
		if current_asset_price >= sell_limit_price {
			// We are in profit already, sell at market price and close the bot.
			fmt.Println(fmt.Sprintf("Current Asset Price: %v is greater than Sell Limit Price: %v. Hence the Bot is in profit already.", current_asset_price, sell_limit_price))
			fmt.Println("Closing the deal at current market price.")
			//goto closeDealAtMarketPrice
			CloseAtMarketPrice(alpacaClient)
			dealEditCompleted <- true
		} else {
			if current_asset_price > sell_stop_price {
				sell_stop_price = current_asset_price
			}
			//Modifying the SELL Order.
			replaceOrderRequest := alpaca.ReplaceOrderRequest{}
			req_qty := decimal.NewFromInt(total_buying_quantity)
			replaceOrderRequest.Qty = &req_qty
			req_lim_price := decimal.NewFromFloat(sell_limit_price)
			replaceOrderRequest.LimitPrice = &req_lim_price
			req_stop_price := decimal.NewFromFloat(sell_stop_price)
			replaceOrderRequest.StopPrice = &req_stop_price
			replaceOrderRequest.TimeInForce = alpaca.GTC
			orderPlaced, err := alpacaClient.ReplaceOrder(orders[1].OrderId, replaceOrderRequest)
			if err != nil {
				for {
					orderPlaced, err = alpacaClient.ReplaceOrder(orders[1].OrderId, replaceOrderRequest)
					if err == nil {
						break
					}
				}
			}
			var order models.OrderMongo
			jsonOrder, _ := json.Marshal(orderPlaced)
			_ = json.Unmarshal(jsonOrder, &order)
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
			order.UserId = user_id
			order.StrategyId = strategy_id
			order.ExchangeId = exchange_id
			order.DealId = deal_id
			order.StrategyName = strategy_name
			for {
				_, err := orderCollection.InsertOne(context.TODO(), order)
				if err == nil {
					break
				}
			}
			oldOrderDetails, _ := alpacaClient.GetOrder(orders[1].OrderId)
			_ = orderCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": orders[1].OrderId}, bson.M{"$set": oldOrderDetails}, options.FindOneAndUpdate().SetReturnDocument(1))
			orders[1].Mu.Lock()
			orders[1].OrderId = order.ID
			orders[1].Mu.Unlock()
		}

	}

	if deal.MaxSafetyTradeCount != max_safety_order_count {
		max_safety_order_count = deal.MaxSafetyTradeCount
	}

	if deal.MaxActiveSafetyTradeCount != max_active_safety_order_count {
		max_active_safety_order_count = deal.MaxActiveSafetyTradeCount
		var countSO int64 = 0
		for _, v := range orders {
			if v.OrderType == "SO" && v.OrderStatus != "filled" {
				countSO++
			}
		}
		if countSO < max_active_safety_order_count {
			// Place more Safety Orders
			fmt.Println("Count of currently active safety orders is less than maximum active safety order count. Hence placing more safety orders.")
			var orders_data models.OrdersData
			var order models.OrderMongo
			for i := countSO; i < max_active_safety_order_count; i++ {
				safety_order_step := previous_safety_order_step_scale * safety_order_step_scale
				safety_order_deviation := previous_safety_order_deviation + safety_order_step
				safety_order_limit_price := previous_safety_order_buying_price * (1.0 - (safety_order_deviation / 100))
				safety_order_stop_price := safety_order_limit_price * 1.02
				if current_asset_price := GetCurrentAssetPrice(alpacaClient); current_asset_price < safety_order_stop_price {
					safety_order_stop_price = current_asset_price
				}
				safety_order_volume := previous_safety_order_volume
				if i != 0 {
					safety_order_volume = safety_order_volume * safety_order_volume_scale
				}
				safety_order_quantity := int64(safety_order_volume / safety_order_limit_price)
				buySafetyOrderPlaced, err := orderpackage.PlaceOrder(alpacaClient, asset, safety_order_quantity, "BUY", "stop_limit", safety_order_limit_price, safety_order_stop_price)
				if err == nil {
					//fmt.Println(fmt.Sprintf("BUY Safety Order Placed: %v ", buySafetyOrderPlaced))
					orders_data.OrderId = buySafetyOrderPlaced.ID
					orders_data.OrderType = "SO"
					orders_data.OrderStatus = "new"
					orders = append(orders, orders_data)
					jsonOrder, _ := json.Marshal(buySafetyOrderPlaced)
					_ = json.Unmarshal(jsonOrder, &order)
					order.Qty, _ = buySafetyOrderPlaced.Qty.Float64()
					order.Notional, _ = buySafetyOrderPlaced.Notional.Float64()
					order.FilledQty, _ = buySafetyOrderPlaced.FilledQty.Float64()

					if buySafetyOrderPlaced.LimitPrice != nil {
						order.LimitPrice, _ = buySafetyOrderPlaced.LimitPrice.Float64()
					}
					if buySafetyOrderPlaced.FilledAvgPrice != nil {
						order.FilledAvgPrice, _ = buySafetyOrderPlaced.FilledAvgPrice.Float64()
					}
					if buySafetyOrderPlaced.StopPrice != nil {
						order.StopPrice, _ = buySafetyOrderPlaced.StopPrice.Float64()
					}
					order.UserId = user_id
					order.StrategyId = strategy_id
					order.ExchangeId = exchange_id
					order.DealId = deal_id
					order.StrategyName = strategy_name
					for {
						_, err = orderCollection.InsertOne(context.TODO(), order)
						if err == nil {
							break
						}
					}
				}
				previous_safety_order_buying_price = safety_order_limit_price
				previous_safety_order_volume = safety_order_volume
				previous_safety_order_deviation = safety_order_deviation
				previous_safety_order_step_scale = safety_order_step
			}

			fmt.Println(fmt.Sprintf("Orders Array: %v", orders))
		} else if countSO > max_active_safety_order_count {
			// Cancel some Safety Orders
			fmt.Println("Count of currently safety orders is greater than maximum active safety order count. Hence cancelling last safety orders.")
			for i := len(orders); countSO != max_active_safety_order_count; i-- {
				fmt.Println(fmt.Sprintf("Cancelling Safety Order: %v", orders[i-1].OrderId))
				CancelOrder(orders[i-1].OrderId, alpacaClient)
				//orders[i-1].Mu.Lock()
				orders = orders[:i-1]
				//orders[i-1].Mu.Unlock()
				countSO--
			}
			fmt.Println(fmt.Sprintf("Orders Array: %v", orders))
		}
	}
	if stopLossPercent, _ := strconv.ParseFloat(deal.StopLossPercent, 64); stopLossPercent != stop_loss_percent {
		stop_loss_percent = stopLossPercent
	}
	//id, _ = primitive.ObjectIDFromHex(deal_id)
	updateDeal["deal_edited_by_user"] = false
	for {
		result := dealsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$set": updateDeal})
		if result.Err() == nil {
			break
		}
	}
	dealEditCompleted <- true

}
