package models

import (
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Stock struct {
	StockName string `json:"stock_name" bson:"stock_name"`
}

type Strategy struct {
	Id                        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StrategyName              string             `json:"strategy_name" bson:"strategy_name"`
	SelectedExchange          string             `json:"selected_exchange" bson:"selected_exchange"`
	SelectedExchangeName      string             `json:"selected_exchange_name" bson:"selected_exchange_name"`
	StrategyType              string             `json:"strategy_type" bson:"strategy_type"`
	StartOrderType            string             `json:"start_order_type" bson:"start_order_type"`
	DealStartCondition        string             `json:"deal_start_condition" bson:"deal_start_condition"`
	BaseOrderSize             float64            `json:"base_order_size" bson:"base_order_size"`
	SafetyOrderSize           float64            `json:"safety_order_size" bson:"safety_order_size"`
	MaxSafetyTradeCount       int64              `json:"max_safety_trade_count" bson:"max_safety_trade_count"`
	MaxActiveSafetyTradeCount int64              `json:"max_active_safety_trade_count" bson:"max_active_safety_trade_count"`
	PriceDevation             string             `json:"price_devation" bson:"price_devation"`
	SafetyOrderVolumeScale    float64            `json:"safety_order_volume_scale" bson:"safety_order_volume_scale"`
	SafetyOrderStepScale      float64            `json:"safety_order_step_scale" bson:"safety_order_step_scale"`
	TakeProfit                string             `json:"take_profit" bson:"take_profit"`
	TargetProfit              string             `json:"target_profit" bson:"target_profit"`
	StopLossPercent           string             `json:"stop_loss_percent" bson:"stop_loss_percent"`
	AllocateFundsToStrategy   string             `json:"allocate_funds_to_strategy" bson:"allocate_funds_to_strategy"`
	UserId                    string             `json:"user_id" bson:"user_id"`
	Version                   int64              `json:"version" bson:"version"`
	Status                    string             `json:"status" bson:"status"`
	TotalDeals                int64              `json:"total_deals" bson:"total_deals"`
	ActiveDeals               int64              `json:"active_deals" bson:"active_deals"`
	Stock                     []*Stock           `json:"stock" bson:"stock"`
}

type StrategyRevision struct {
	Id                        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StrategyName              string             `json:"strategy_name" bson:"strategy_name"`
	SelectedExchange          string             `json:"selected_exchange" bson:"selected_exchange"`
	SelectedExchangeName      string             `json:"selected_exchange_name" bson:"selected_exchange_name"`
	StrategyType              string             `json:"strategy_type" bson:"strategy_type"`
	StartOrderType            string             `json:"start_order_type" bson:"start_order_type"`
	DealStartCondition        string             `json:"deal_start_condition" bson:"deal_start_condition"`
	BaseOrderSize             float64            `json:"base_order_size" bson:"base_order_size"`
	SafetyOrderSize           float64            `json:"safety_order_size" bson:"safety_order_size"`
	MaxSafetyTradeCount       int64              `json:"max_safety_trade_count" bson:"max_safety_trade_count"`
	MaxActiveSafetyTradeCount int64              `json:"max_active_safety_trade_count" bson:"max_active_safety_trade_count"`
	PriceDevation             string             `json:"price_devation" bson:"price_devation"`
	SafetyOrderVolumeScale    float64            `json:"safety_order_volume_scale" bson:"safety_order_volume_scale"`
	SafetyOrderStepScale      float64            `json:"safety_order_step_scale" bson:"safety_order_step_scale"`
	TakeProfit                string             `json:"take_profit" bson:"take_profit"`
	TargetProfit              string             `json:"target_profit" bson:"target_profit"`
	StopLossPercent           string             `json:"stop_loss_percent" bson:"stop_loss_percent"`
	AllocateFundsToStrategy   string             `json:"allocate_funds_to_strategy" bson:"allocate_funds_to_strategy"`
	UserId                    string             `json:"user_id" bson:"user_id"`
	Version                   int64              `json:"version" bson:"version"`
	Status                    string             `json:"status" bson:"status"`
	TotalDeals                int64              `json:"total_deals" bson:"total_deals"`
	ActiveDeals               int64              `json:"active_deals" bson:"active_deals"`
	Stock                     []*Stock           `json:"stock" bson:"stock"`
	StrategyId                string             `json:"strategy_id" bson:"strategy_id"`
}

type Deal struct {
	Id                            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StrategyId                    string             `json:"strategy_id" bson:"strategy_id"`
	StrategyVersion               int64              `json:"strategy_version" bson:"strategy_version"`
	Stock                         string             `json:"stock" bson:"stock"`
	UserId                        string             `json:"user_id" bson:"user_id"`
	Status                        string             `json:"status" bson:"status"`
	MaxActiveSafetyTradeCount     int64              `json:"max_active_safety_trade_count" bson:"max_active_safety_trade_count"`
	MaxSafetyTradeCount           int64              `json:"max_safety_trade_count" bson:"max_safety_trade_count"`
	ActiveSafetyOrderCount        int64              `json:"active_safety_order_count" bson:"active_safety_order_count"`
	FilledSafetyOrderCount        int64              `json:"filled_safety_order_count" bson:"filled_safety_order_count"`
	CreatedAt                     time.Time          `json:"created_at" bson:"created_at"`
	ClosedAt                      time.Time          `json:"closed_at" bson:"closed_at"`
	TotalOrderQuantity            int64              `json:"total_order_quantity" bson:"total_order_quantity"`
	ProfitPercentage              string             `json:"profit_percentage" bson:"profit_percentage"`
	ProfitValue                   float64            `json:"profit_value" bson:"profit_value"`
	TotalBuyingPrice              float64            `json:"total_buying_price" bson:"total_buying_price"`
	TotalSellPrice                float64            `json:"total_sell_price" bson:"total_sell_price"`
	TargetProfit                  string             `json:"target_profit" bson:"target_profit"`
	StrategyName                  string             `json:"strategy_name" bson:"strategy_name"`
	SelectedExchange              string             `json:"selected_exchange" bson:"selected_exchange"`
	BaseOrderSize                 float64            `json:"base_order_size" bson:"base_order_size"`
	SafetyOrderSize               float64            `json:"safety_order_size" bson:"safety_order_size"`
	SafetyOrderVolumeScale        float64            `json:"safety_order_volume_scale" bson:"safety_order_volume_scale"`
	DealCancelledByUser           bool               `json:"deal_cancelled_by_user" bson:"deal_cancelled_by_user"`
	DealClosedAtMarketPriceByUser bool               `json:"deal_closed_at_market_price_by_user" bson:"deal_closed_at_market_price_by_user"`
	StopLossPercent               string             `json:"stop_loss_percent" bson:"stop_loss_percent"`
	DealEditedByUser              bool               `json:"deal_edited_by_user" bson:"deal_edited_by_user"`
	ManualOrderPlacedByUser       bool               `json:"manual_order_placed_by_user" bson:"manual_order_placed_by_user"`
	NextSafetyOrderLimitPrice     float64            `json:"next_safety_order_limit_price" bson:"next_safety_order_limit_price"`
	AvgBuyingPrice                float64            `json:"avg_buying_price" bson:"avg_buying_price"`
	//TotalOrderAmount          float64            `json:"total_order_amount" bson:"total_order_amount"`
}
type DealRequest struct {
	StrategyId string   `json:"strategy_id" bson:"strategy_id"`
	Stock      *[]Stock `json:"stock" bson:"stock"`
}

type EventHistory struct {
	Id            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	OperationType string             `json:"operation_type" bson:"operation_type"`
	Timestamp     string             `json:"timestamp" bson:"timestamp"`
	Db            string             `json:"db" bson:"db"`
	Collection    string             `json:"collection" bson:"collection"`
	Name          string             `json:"name" bson:"name"`
	UserId        string             `json:"user_id" bson:"user_id"`
	StrategyId    string             `json:"strategy_id" bson:"strategy_id"`
	OldValue      Strategy           `json:"old_value" bson:"old_value"`
	NewValue      Strategy           `json:"new_value" bson:"new_value"`
}

type DealEventHistory struct {
	Id            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	OperationType string             `json:"operation_type" bson:"operation_type"`
	Timestamp     string             `json:"timestamp" bson:"timestamp"`
	Db            string             `json:"db" bson:"db"`
	Collection    string             `json:"collection" bson:"collection"`
	Name          string             `json:"name" bson:"name"`
	UserId        string             `json:"user_id" bson:"user_id"`
	StrategyId    string             `json:"strategy_id" bson:"strategy_id"`
	DealId        string             `json:"deal_id" bson:"deal_id"`
	OldValue      Deal               `json:"old_value" bson:"old_value"`
	NewValue      Deal               `json:"new_value" bson:"new_value"`
}

type AccountInfo struct {
	Balance                      decimal.Decimal `json:"balance"`
	MaxAmtStrategyUsage          decimal.Decimal `json:"max_amt_strategy_usage"`
	MaxSafetyOrderPriceDeviation string          `json:"max_safety_order_price_deviation"`
	AvailableBalance             float64         `json:"available_balance"`
	BuyingPower                  decimal.Decimal `json:"buying_power`
	Equity                       decimal.Decimal `json:"equity"`
}

type Exchange struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SelectedExchange string             `json:"selected_exchange" bson:"selected_exchange"`
	ExchangeName     string             `json:"exchange_name" bson:"exchange_name"`
	ExchangeType     string             `json:"exchange_type" bson:"exchange_type"`
	UserId           string             `json:"user_id" bson:"user_id"`
	ApiKey           string             `json:"api_key" bson:"api_key"`
	ApiSecret        string             `json:"api_secret" bson:"api_secret"`
	Active           *bool              `json:"active" bson:"active"`
}

type ErrorString struct {
	S string
}

func (e *ErrorString) Error() string {
	return e.S
}

type CancelDealResponse struct {
	Cancelled bool `json:"cancelled"`
}

type DealJson struct {
	DealId string `json:"deal_id"`
	Asset  string `json:"asset"`
}

type ManipulateDeal struct {
	Status                    string `json:"status"`
	MaxActiveSafetyTradeCount int64  `json:"max_active_safety_trade_count" bson:"max_active_safety_trade_count"`
	MaxSafetyTradeCount       int64  `json:"max_safety_trade_count" bson:"max_safety_trade_count"`
	TargetProfit              string `json:"target_profit" bson:"target_profit"`
	StopLossPercent           string `json:"stop_loss_percent" bson:"stop_loss_percent"`
}

type BuyMoreRequest struct {
	DealId string          `json:"deal_id" bson:"deal_id"`
	Symbol string          `json:"symbol" bson:"symbol"`
	Qty    decimal.Decimal `json:"qty" bson:"qty"`
	Type   OrderType       `json:"order_type" bson:"order_type"`
}

type Order struct {
	ID             string           `json:"id,omitempty" bson:"_id,omitempty"`
	ClientOrderId  string           `json:"client_order_id" bson:"client_order_id"`
	CreatedAt      time.Time        `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" bson:"updated_at"`
	SubmittedAt    time.Time        `json:"submitted_at" bson:"submitted_at"`
	FilledAt       *time.Time       `json:"filled_at" bson:"filled_at"`
	ExpiredAt      *time.Time       `json:"expired_at" bson:"expired_at"`
	CanceledAt     *time.Time       `json:"canceled_at" bson:"canceled_at"`
	FailedAt       *time.Time       `json:"failed_at" bson:"failed_at"`
	ReplacedAt     *time.Time       `json:"replaced_at" bson:"replaced_at"`
	Replaces       *string          `json:"replaces" bson:"replaces"`
	ReplacedBy     *string          `json:"replaced_by" bson:"replaced_by"`
	AssetId        string           `json:"asset_id" bson:"asset_id"`
	Symbol         string           `json:"symbol" bson:"symbol"`
	Exchange       string           `json:"exchange" bson:"exchange"`
	Class          string           `json:"asset_class" bson:"asset_class"`
	Qty            decimal.Decimal  `json:"qty" bson:"qty"`
	Notional       decimal.Decimal  `json:"notional" bson:"notional"`
	FilledQty      decimal.Decimal  `json:"filled_qty" bson:"filled_qty"`
	Type           OrderType        `json:"order_type" bson:"order_type"`
	Side           Side             `json:"side" bson:"side"`
	TimeInForce    TimeInForce      `json:"time_in_force" bson:"time_in_force"`
	LimitPrice     *decimal.Decimal `json:"limit_price" bson:"limit_price"`
	FilledAvgPrice *decimal.Decimal `json:"filled_avg_price" bson:"filled_avg_price"`
	StopPrice      *decimal.Decimal `json:"stop_price" bson:"stop_price"`
	TrailPrice     *decimal.Decimal `json:"trail_price" bson:"trail_price"`
	TrailPercent   *decimal.Decimal `json:"trail_percent" bson:"trail_percent"`
	Hwm            *decimal.Decimal `json:"hwm" bson:"hwm"`
	Status         string           `json:"status" bson:"status"`
	ExtendedHours  bool             `json:"extended_hours" bson:"extended_hours"`
	Legs           *[]Order         `json:"legs" bson:"legs"`
	UserId         string           `json:"user_id" bson:"user_id"`
	ExchangeId     string           `json:"exchange_id" bson:"exchange_id"`
	StrategyId     string           `json:"strategy_id" bson:"strategy_id"`
	DealId         string           `json:"deal_id" bson:"deal_id"`
	StrategyName   string           `json:"strategy_name" bson:"strategy_name"`
}
type OrderType string

const (
	Market       OrderType = "market"
	Limit        OrderType = "limit"
	Stop         OrderType = "stop"
	StopLimit    OrderType = "stop_limit"
	TrailingStop OrderType = "trailing_stop"
)

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

type TimeInForce string

const (
	Day TimeInForce = "day"
	GTC TimeInForce = "gtc"
	OPG TimeInForce = "opg"
	IOC TimeInForce = "ioc"
	FOK TimeInForce = "fok"
	GTX TimeInForce = "gtx"
	GTD TimeInForce = "gtd"
	CLS TimeInForce = "cls"
)

type AccountProfit struct {
	TodaysProfit    float64 `json:"todays_profit" bson:"todays_profit"`
	TotalProfit     float64 `json:"total_profit" bson:"total_profit"`
	ActiveDeals     int64   `json:"active_deals" bson:"active_deals"`
	CompletedProfit float64 `json:"completed_profit" bson:"completed_profit"`
}

type Strategy_Profits struct {
	StrategyId  string    `json:"strategy_id" bson:"strategy_id"`
	UserId      string    `json:"user_id" bson:"user_id"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	ProfitValue float64   `json:"profit_value" bson:"profit_value"`
}

type Profits_Response struct {
	Timestamp [5]int64   `json:"timestamps"`
	Profits   [5]float64 `json:"profits"`
}
