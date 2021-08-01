package models

import (
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

type OrdersData struct {
	OrderId     string
	OrderType   string
	OrderStatus string
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
	TotalOrderQuantity            int64              `json:"total_order_quantity" bson:"total_order_quantity"`
	ProfitPercentage              string             `json:"profit_percentage" bson:"profit_percentage"`
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