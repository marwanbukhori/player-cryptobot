package database

import (
	"database/sql"
	"fmt"

	"github.com/marwanbukhori/player-cryptobot/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database wraps the underlying database connection and provides methods for data access
type Database struct {
	db   *sql.DB
	gorm *gorm.DB
}

// Initialize creates a new database connection and sets up the schema
func Initialize(dsn string) (*Database, error) {
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := initSchema(gormDB); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB: %v", err)
	}

	return &Database{
		db:   sqlDB,
		gorm: gormDB,
	}, nil
}

// SaveTrade saves a trade to the database
func (db *Database) SaveTrade(trade *models.Trade) error {
	return db.gorm.Create(trade).Error
}

// GetTradingSummary returns a summary of all trades
func (db *Database) GetTradingSummary() ([]models.TradingSummary, error) {
	var summaries []models.TradingSummary
	err := db.gorm.Raw(`
		SELECT
			symbol,
			COUNT(*) as total_trades,
			SUM(CASE WHEN pnl > 0 THEN 1 ELSE 0 END) as winning_trades,
			SUM(CASE WHEN pnl < 0 THEN 1 ELSE 0 END) as losing_trades,
			ROUND(SUM(pnl), 4) as total_pnl,
			ROUND(AVG(pnl_percent), 2) as avg_pnl_percent,
			ROUND(SUM(value), 4) as total_volume,
			MIN(timestamp) as first_trade,
			MAX(timestamp) as last_trade
		FROM trades
		GROUP BY symbol`).Scan(&summaries).Error
	return summaries, err
}

// GetLastBuyTrade returns the last buy trade for a given symbol
func (db *Database) GetLastBuyTrade(symbol string) (*models.Trade, error) {
	var trade models.Trade
	err := db.gorm.Where("symbol = ? AND side = ?", symbol, "BUY").
		Order("timestamp DESC").
		First(&trade).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no previous buy trade found")
		}
		return nil, err
	}
	return &trade, nil
}

// GetAllTrades returns all trades
func (db *Database) GetAllTrades() ([]models.Trade, error) {
	var trades []models.Trade
	err := db.gorm.Order("timestamp DESC").Find(&trades).Error
	return trades, err
}

// GetRecentTrades returns the most recent trades up to the specified limit
func (db *Database) GetRecentTrades(limit int) ([]models.Trade, error) {
	var trades []models.Trade
	err := db.gorm.Order("timestamp DESC").Limit(limit).Find(&trades).Error
	return trades, err
}

// GetOpenPositions returns all open positions
func (db *Database) GetOpenPositions() ([]models.Trade, error) {
	var trades []models.Trade
	err := db.gorm.Where("status = ?", "OPEN").Find(&trades).Error
	return trades, err
}

// CalculateOpenPnl calculates the unrealized PnL for all open positions
func (db *Database) CalculateOpenPnl(currentPrice float64) (float64, error) {
	openPositions, err := db.GetOpenPositions()
	if err != nil {
		return 0, err
	}

	var openPnl float64
	for _, pos := range openPositions {
		openPnl += (currentPrice - pos.Price) * pos.Quantity
	}
	return openPnl, nil
}

func initSchema(db *gorm.DB) error {
	// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes
	return db.AutoMigrate(&models.Trade{})
}
