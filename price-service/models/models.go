package models

import (
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Exchange struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SelectedExchange string             `json:"selected_exchange" bson:"selected_exchange"`
	ExchangeName     string             `json:"exchange_name" bson:"exchange_name"`
	ExchangeType     string             `json:"exchange_type" bson:"exchange_type"`
	UserId           string             `json:"user_id" bson:"user_id"`
	ApiKey           string             `json:"api_key" bson:"api_key"`
	ApiSecret        string             `json:"api_secret" bson:"api_secret"`
}

type Position struct {
	AssetID        string          `json:"asset_id"`
	Symbol         string          `json:"symbol"`
	Exchange       string          `json:"exchange"`
	Class          string          `json:"asset_class"`
	AccountID      string          `json:"account_id"`
	EntryPrice     decimal.Decimal `json:"avg_entry_price"`
	Qty            decimal.Decimal `json:"qty"`
	Side           string          `json:"side"`
	MarketValue    decimal.Decimal `json:"market_value"`
	CostBasis      decimal.Decimal `json:"cost_basis"`
	UnrealizedPL   decimal.Decimal `json:"unrealized_pl"`
	UnrealizedPLPC decimal.Decimal `json:"unrealized_plpc"`
	CurrentPrice   decimal.Decimal `json:"current_price"`
	LastdayPrice   decimal.Decimal `json:"lastday_price"`
	ChangeToday    decimal.Decimal `json:"change_today"`
}

type Asset struct {
	Id           string `json:"_id,omitempty" bson:"_id,omitempty"`
	Name         string `json:"name" bson:"name"`
	Exchange     string `json:"exchange" bson:"exchange"`
	AssetClass   string `json:"asset_class" bson:"asset_class"`
	Symbol       string `json:"symbol" bson:"symbol"`
	Status       string `json:"status" bson:"status"`
	Tradable     bool   `json:"tradable" bson:"tradable"`
	Marginable   bool   `json:"marginable" bson:"marginable"`
	Shortable    bool   `json:"shortable bson:"shortable"`
	EasyToBorrow bool   `json:"easy_to_borrow" bson:"easy_to_borrow"`
}

type ErrorString struct {
	S string
}

func (e *ErrorString) Error() string {
	return e.S
}

type PriceResponse struct {
	CurrentPrice float64 `json:"current_price" bson:"current_price"`
}
