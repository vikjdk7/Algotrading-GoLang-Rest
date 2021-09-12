package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/gorilla/mux"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/dashboard-service/helper"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/dashboard-service/middleware"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/dashboard-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getPortfolioHistory(w http.ResponseWriter, r *http.Request) {
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

	var period string
	var timeframe alpaca.RangeFreq
	var extendedHours bool

	graphPeriod := r.URL.Query().Get("period")

	if graphPeriod == "" {
		customError.S = "Period missing in query parameters. Please use one of [day, month, year]."
		helper.GetError(&customError, w)
		return
	}

	if graphPeriod != "day" && graphPeriod != "month" && graphPeriod != "year" {
		customError.S = "Invalid value of period in query parameters. Allowed values: [day, month, year]."
		helper.GetError(&customError, w)
		return
	}

	if graphPeriod == "day" {
		period = "1D"
		timeframe = alpaca.Min5
	} else if graphPeriod == "month" {
		period = "1M"
		timeframe = alpaca.Day1
	} else if graphPeriod == "year" {
		period = "1A"
		timeframe = alpaca.Day1
	}

	exchangeId := r.URL.Query().Get("exchange-id")
	var exchange models.Exchange

	if exchangeId == "" {
		customError.S = "Exchange ID is missing in query parameters"
		helper.GetError(&customError, w)
		return
	} else { //Exchange Id sent by user
		id, _ := primitive.ObjectIDFromHex(exchangeId)
		err := exchangeCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId}).Decode(&exchange)
		if err != nil {
			helper.GetError(err, w)
			return
		}
	}
	os.Setenv(common.EnvApiKeyID, exchange.ApiKey)
	os.Setenv(common.EnvApiSecretKey, exchange.ApiSecret)
	if exchange.ExchangeType == "paper_trading" {
		alpaca.SetBaseUrl("https://paper-api.alpaca.markets")
	} else if exchange.ExchangeType == "live_trading" {
		alpaca.SetBaseUrl("https://api.alpaca.markets")
	}

	alpacaClient := alpaca.NewClient(common.Credentials())

	fmt.Println(fmt.Sprintf("New Request with Period: %v, Timeframe: %v, User: %v, exchangeID: %v", period, timeframe, userId, exchangeId))
	portfolioHistory, err := alpacaClient.GetPortfolioHistory(&period, &timeframe, nil, extendedHours)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(portfolioHistory)
}

func getAccountInfo(w http.ResponseWriter, r *http.Request) {
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

	exchangeId := r.URL.Query().Get("exchange-id")
	var exchange models.Exchange

	if exchangeId == "" {
		customError.S = "Exchange ID is missing in query parameters"
		helper.GetError(&customError, w)
		return
	} else { //Exchange Id sent by user
		id, _ := primitive.ObjectIDFromHex(exchangeId)
		err := exchangeCollection.FindOne(context.TODO(), bson.M{"_id": id, "user_id": userId}).Decode(&exchange)
		if err != nil {
			helper.GetError(err, w)
			return
		}
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

	var accountinfo models.AccountInfo

	accountinfo.BuyingPower = acct.BuyingPower.Round(2)
	fmt.Println("Buying Power rounded to 2: ", acct.BuyingPower.Round(2))
	accountinfo.Equity = acct.Equity.Round(2)
	accountinfo.Cash = acct.Cash.Round(2)
	accountinfo.LongMarketValue = acct.LongMarketValue.Round(2)

	var todays_profit float64
	var total_profit float64

	statuses := [3]string{"bought", "completed", "cancelled"}
	cur, err := dealsCollection.Find(context.TODO(), bson.M{"user_id": userId, "status": bson.M{"$in": statuses}})
	if err != nil {
		helper.GetError(err, w)
		return
	}
	defer cur.Close(context.TODO())
	currentTimestamp := time.Now()
	for cur.Next(context.TODO()) {
		var deal models.Deal
		err := cur.Decode(&deal) // decode similar to deserialize process.
		if err != nil {
			helper.GetError(err, w)
			return
		}
		if deal.Status == "bought" {
			currentAssetPrice := GetCurrentAssetPrice(deal.Stock, alpacaClient)
			todays_profit += ((currentAssetPrice * float64(deal.TotalOrderQuantity)) - deal.TotalBuyingPrice)
			total_profit += ((currentAssetPrice * float64(deal.TotalOrderQuantity)) - deal.TotalBuyingPrice)

		} else if deal.Status == "cancelled" {
			if DateEqual(currentTimestamp, deal.ClosedAt) == true {
				todays_profit += deal.ProfitValue
			}
			total_profit += deal.ProfitValue
		} else if deal.Status == "completed" {
			if DateEqual(currentTimestamp, deal.ClosedAt) == true {
				todays_profit += deal.ProfitValue
			}
			total_profit += deal.ProfitValue
		}
	}

	accountinfo.TotalProfitLoss = total_profit
	accountinfo.TodayProfitLoss = todays_profit
	todProf, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", todays_profit), 64)
	accountinfo.TodayProfitLoss = todProf
	totProf, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", total_profit), 64)
	accountinfo.TotalProfitLoss = totProf
	json.NewEncoder(w).Encode(accountinfo)
}

func GetCurrentAssetPrice(asset string, alpacaClient *alpaca.Client) float64 {
	latestQuote, err := alpacaClient.GetLatestQuote(asset)
	if err != nil {
		for {
			latestQuote, err = alpacaClient.GetLatestQuote(asset)
			if err == nil && latestQuote.AskPrice != 0.0 {
				break
			}
		}
	}
	return latestQuote.AskPrice
}
func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

var exchangeCollection *mongo.Collection
var dealsCollection *mongo.Collection

func init() {
	//Connect to mongoDB with helper class
	exchangeCollection, dealsCollection = helper.ConnectDB()
}

func main() {
	//Init Router
	r := mux.NewRouter()

	// arrange our route
	r.HandleFunc("/DashboardService/api/v1/portfolio/history", getPortfolioHistory).Methods("GET")
	r.HandleFunc("/DashboardService/api/v1/portfolio/info", getAccountInfo).Methods("GET")

	// set our port address
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
