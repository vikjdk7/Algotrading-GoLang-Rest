package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

type Strategy_Profits struct {
	StrategyId  string    `json:"strategy_id" bson:"strategy_id"`
	UserId      string    `json:"user_id" bson:"user_id"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	ProfitValue float64   `json:"profit_value" bson:"profit_value"`
}
