package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Stock struct {
	StockName string `json:"stock_name" bson:"stock_name"`
}

type Strategy struct {
	Id                        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StrategyName              string             `json:"strategy_name,omitempty" bson:"strategy_name"`
	SelectedExchange          string             `json:"selected_exchange,omitempty" bson:"selected_exchange"`
	StrategyType              string             `json:"strategy_type,omitempty" bson:"strategy_type"`
	StartOrderType            string             `json:"start_order_type,omitempty" bson:"start_order_type"`
	DealStartCondition        string             `json:"deal_start_condition,omitempty" bson:"deal_start_condition"`
	BaseOrderSize             float64            `json:"base_order_size,omitempty" bson:"base_order_size"`
	SafetyOrderSize           float64            `json:"safety_order_size,omitempty" bson:"safety_order_size"`
	MaxSafetyTradeCount       int64              `json:"max_safety_trade_count,omitempty" bson:"max_safety_trade_count"`
	MaxActiveSafetyTradeCount int64              `json:"max_active_safety_trade_count,omitempty" bson:"max_active_safety_trade_count"`
	PriceDevation             string             `json:"price_devation,omitempty" bson:"price_devation"`
	SafetyOrderVolumeScale    float64            `json:"safety_order_volume_scale,omitempty" bson:"safety_order_volume_scale"`
	SafetyOrderStepScale      float64            `json:"safety_order_step_scale,omitempty" bson:"safety_order_step_scale"`
	TakeProfit                string             `json:"take_profit,omitempty" bson:"take_profit"`
	TargetProfit              string             `json:"target_profit,omitempty" bson:"target_profit"`
	StopLossPercent           string             `json:"stop_loss_percent,omitempty" bson:"stop_loss_percent"`
	AllocateFundsToStrategy   string             `json:"allocate_funds_to_strategy,omitempty" bson:"allocate_funds_to_strategy"`
	UserId                    string             `json:"user_id,omitempty" bson:"user_id"`
	Version                   int64              `json:"version,omitempty" bson:"version"`
	Status                    string             `json:"status,omitempty" bson:"status"`
	Stock                     []*Stock           `json:"stock,omitempty" bson:"stock"`
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
	Id                        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StrategyId                string             `json:"strategy_id,omitempty" bson:"strategy_id"`
	Stock                     string             `json:"stock,omitempty" bson:"stock"`
	UserId                    string             `json:"user_id,omitempty" bson:"user_id"`
	Status                    string             `json:"status,omitempty" bson:"status"`
	MaxActiveSafetyTradeCount int64              `json:"max_active_safety_trade_count,omitempty" bson:"max_active_safety_trade_count"`
	MaxSafetyTradeCount       int64              `json:"max_safety_trade_count,omitempty" bson:"max_safety_trade_count"`
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

type EventHistoryResponse struct {
	Strategy *[]EventHistory     `json:"strategies"`
	Deal     *[]DealEventHistory `json:"deals"`
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
