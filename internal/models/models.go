package models

import "time"

type MarketData struct {
	Symbol string
	Price  float64
	Time   time.Time
}

type Signal struct {
	Symbol    string
	Action    string // "buy" or "sell"
	Price     float64
	Quantity  float64
	Timestamp time.Time
}

type Order struct {
	Symbol    string
	Side      string // "BUY" or "SELL"
	Type      string // "MARKET" or "LIMIT"
	Quantity  float64
	Price     float64
	Timestamp time.Time
	Status    string
}

type Kline struct {
	OpenTime  int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime int64
}
