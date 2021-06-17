package models

import (
	"time"

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

type EventHistory struct {
	Id            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	OperationType string             `json:"operation_type" bson:"operation_type"`
	Timestamp     string             `json:"timestamp" bson:"timestamp"`
	Db            string             `json:"db" bson:"db"`
	Collection    string             `json:"collection" bson:"collection"`
	Name          string             `json:"name" bson:"name"`
	UserId        string             `json:"user_id" bson:"user_id"`
	ExchangeId    string             `json:"exchange_id" bson:"exchange_id"`
	OldValue      Exchange           `json:"old_value" bson:"old_value"`
	NewValue      Exchange           `json:"new_value" bson:"new_value"`
}

type ExchangeAccountInfo struct {
	Id                    string          `json:"id"`
	AccountNumber         string          `json:"account_number"`
	CreatedAt             time.Time       `json:"created_at"`
	Status                string          `json:"status"`
	Currency              string          `json:"currency"`
	Cash                  decimal.Decimal `json:"cash"`
	CashWithdrawable      decimal.Decimal `json:"cash_withdrawable"`
	TradingBlocked        bool            `json:"trading_blocked"`
	TransfersBlocked      bool            `json:"transfers_blocked"`
	AccountBlocked        bool            `json:"account_blocked"`
	BuyingPower           decimal.Decimal `json:"buying_power"`
	PatternDayTrader      bool            `json:"pattern_day_trader"`
	DaytradeCount         int64           `json:"daytrade_count"`
	DaytradingBuyingPower decimal.Decimal `json:"daytrading_buying_power"`
	RegtBuyingPower       decimal.Decimal `json:"regt_buying_power"`
	Equity                decimal.Decimal `json:"equity"`
	LastEquity            decimal.Decimal `json:"last_equity"`
	InitialMargin         decimal.Decimal `json:"initial_margin"`
	LongMarketValue       decimal.Decimal `json:"long_market_value"`
	ShortMarketValue      decimal.Decimal `json:"short_market_value"`
}

type CreateExchangeResponse struct {
	Exchange    Exchange            `json:"exchange"`
	AccountInfo ExchangeAccountInfo `json:"account_info"`
}

type ErrorString struct {
	S string
}

func (e *ErrorString) Error() string {
	return e.S
}
