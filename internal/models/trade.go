package models

import "time"

type Trade struct {
	ID         string `gorm:"primaryKey"`
	PositionID string
	Symbol     string
	Side       string
	Price      float64
	Quantity   float64
	Value      float64
	Fee        float64
	Timestamp  time.Time
	PnL        float64
	PnLPercent float64
	CreatedAt  time.Time
	Status     string
}

type TradingSummary struct {
	Symbol        string
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	TotalPnL      float64
	AvgPnLPercent float64
	TotalVolume   float64
	FirstTrade    time.Time
	LastTrade     time.Time
}
