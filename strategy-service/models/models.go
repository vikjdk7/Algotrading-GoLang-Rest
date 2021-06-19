package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Stock struct {
	StockName string `json:"stock_name" bson:"stock_name"`
}

type Strategy struct {
	Id                        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StrategyName              string             `json:"strategy_name" bson:"strategy_name"`
	SelectedExchange          string             `json:"selected_exchange" bson:"selected_exchange"`
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
	Stock                     []*Stock           `json:"stock" bson:"stock"`
}

type StrategyRevision struct {
	Id                        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StrategyName              string             `json:"strategy_name" bson:"strategy_name"`
	SelectedExchange          string             `json:"selected_exchange" bson:"selected_exchange"`
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
	Stock                     []*Stock           `json:"stock" bson:"stock"`
	StrategyId                string             `json:"strategy_id" bson:"strategy_id"`
}

type Deal struct {
	Id         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StrategyId string             `json:"strategy_id" bson:"strategy_id"`
	Version    int64              `json:"version" bson:"version"`
	Stock      string             `json:"stock" bson:"stock"`
	UserId     string             `json:"user_id" bson:"user_id"`
	Status     string             `json:"status" bson:"status"`
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

type ErrorString struct {
	S string
}

func (e *ErrorString) Error() string {
	return e.S
}
