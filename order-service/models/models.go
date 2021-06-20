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
}

type OrderRequest struct {
	ExchangeId  string          `json:"exchange_id" bson:"exchange_id"`
	Symbol      string          `json:"symbol" bson:"symbol"`
	Qty         decimal.Decimal `json:"qty" bson:"qty"`
	Side        Side            `json:"side" bson:"side"`
	Type        OrderType       `json:"order_type" bson:"order_type"`
	TimeInForce TimeInForce     `json:"time_in_force" bson:"time_in_force"`
	LimitPrice  decimal.Decimal `json:"limit_price" bson:"limit_price"`
	StopPrice   decimal.Decimal `json:"stop_price" bson:"stop_price"`
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

type ErrorString struct {
	S string
}

func (e *ErrorString) Error() string {
	return e.S
}
