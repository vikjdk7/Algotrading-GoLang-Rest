package orders

import (
	"errors"
	"fmt"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/shopspring/decimal"
)

func PlaceOrder(alpacaClient *alpaca.Client, asset string, qty int64, orderSide string, orderType string, limitPrice float64, stopPrice float64) (*alpaca.Order, error) {
	placeOrderRequest := alpaca.PlaceOrderRequest{}
	placeOrderRequest.AssetKey = &asset
	placeOrderRequest.Qty = decimal.NewFromInt(qty)
	if orderSide == "BUY" {
		placeOrderRequest.Side = alpaca.Buy
	} else if orderSide == "SELL" {
		placeOrderRequest.Side = alpaca.Sell
	}

	if orderType == "market" {
		placeOrderRequest.Type = alpaca.Market
		placeOrderRequest.TimeInForce = alpaca.Day
	} else if orderType == "stop_limit" {
		placeOrderRequest.Type = alpaca.StopLimit
		limitPriceReq := decimal.NewFromFloat(limitPrice)
		placeOrderRequest.LimitPrice = &limitPriceReq
		stopPriceReq := decimal.NewFromFloat(stopPrice)
		placeOrderRequest.StopPrice = &stopPriceReq
		placeOrderRequest.TimeInForce = alpaca.GTC
	}

	fmt.Print("Placing Order: ")
	fmt.Println(placeOrderRequest)
	orderPlaced, err := alpacaClient.PlaceOrder(placeOrderRequest)
	// RETRY
	if err != nil {
		for {
			orderPlaced, err = alpacaClient.PlaceOrder(placeOrderRequest)
			if err == nil {
				break
			}
		}
	}
	fmt.Println(fmt.Sprintf("orderPlaced: %v", orderPlaced))
	return orderPlaced, nil
}

func GetOrder(alpacaClient *alpaca.Client, orderID string) (string, error) {
	orderDetails, err := alpacaClient.GetOrder(orderID)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not get the order %v", err))
	}
	return orderDetails.Status, nil
}
