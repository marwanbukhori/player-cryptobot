package models

import (
	"time"

	"gorm.io/gorm"
)

// Trade represents a trading transaction
type Trade struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	PositionID string    `gorm:"index;type:varchar(100)"`
	Symbol     string    `gorm:"index;type:varchar(20);not null"`
	Side       string    `gorm:"index;type:varchar(10);not null"` // BUY or SELL
	Price      float64   `gorm:"type:decimal(20,8);not null"`
	Quantity   float64   `gorm:"type:decimal(20,8);not null"`
	Value      float64   `gorm:"type:decimal(20,8);not null"`                      // Price * Quantity
	Fee        float64   `gorm:"type:decimal(20,8);default:0"`                     // Trading fee
	Timestamp  time.Time `gorm:"index;not null"`                                   // When the trade occurred
	PnL        float64   `gorm:"column:pn_l;type:decimal(20,8);default:0"`         // Profit/Loss in USDT
	PnLPercent float64   `gorm:"column:pn_l_percent;type:decimal(10,4);default:0"` // Profit/Loss percentage
	Status     string    `gorm:"index;type:varchar(20);default:'OPEN'"`            // OPEN or CLOSED
	CreatedAt  time.Time `gorm:"autoCreateTime"`                                   // When the record was created
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`                                   // When the record was last updated
}

// TradingSummary represents aggregated trading statistics
type TradingSummary struct {
	Symbol        string    `json:"symbol"`
	TotalTrades   int       `json:"total_trades"`
	WinningTrades int       `json:"winning_trades"`
	LosingTrades  int       `json:"losing_trades"`
	TotalPnL      float64   `json:"total_pnl"`
	AvgPnLPercent float64   `json:"avg_pn_l_percent"`
	TotalVolume   float64   `json:"total_volume"`
	FirstTrade    time.Time `json:"first_trade"`
	LastTrade     time.Time `json:"last_trade"`
}

// TableName specifies the table name for the Trade model
func (Trade) TableName() string {
	return "trades"
}

// BeforeCreate is called before inserting a new trade record
func (t *Trade) BeforeCreate(tx *gorm.DB) error {
	if t.Value == 0 {
		t.Value = t.Price * t.Quantity
	}
	if t.Status == "" {
		t.Status = "OPEN"
	}
	return nil
}
