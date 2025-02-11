package database

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/marwanbukhori/player-cryptobot/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps the underlying database connection and provides methods for data access
type Database struct {
	db   *sql.DB
	gorm *gorm.DB
}

// Initialize creates a new database connection and sets up the schema
func Initialize(dsn string) (*Database, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dsn), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// Configure GORM to be less verbose
	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(io.Discard, "", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Error,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	}

	gormDB, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Get the underlying *sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB: %v", err)
	}

	// Create the database instance
	db := &Database{
		db:   sqlDB,
		gorm: gormDB,
	}

	// Only create table if it doesn't exist
	if !db.gorm.Migrator().HasTable(&models.Trade{}) {
		if err := db.gorm.AutoMigrate(&models.Trade{}); err != nil {
			return nil, fmt.Errorf("failed to create trades table: %v", err)
		}

		// Create indexes only for new database
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_trades_timestamp ON trades(timestamp)",
			"CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol)",
			"CREATE INDEX IF NOT EXISTS idx_trades_side ON trades(side)",
			"CREATE INDEX IF NOT EXISTS idx_trades_status ON trades(status)",
		}

		for _, idx := range indexes {
			if err := db.gorm.Exec(idx).Error; err != nil {
				return nil, fmt.Errorf("failed to create index: %v", err)
			}
		}

		// Create view only for new database
		if err := db.gorm.Exec(`
			CREATE VIEW IF NOT EXISTS trade_summary AS
			SELECT
				symbol,
				COUNT(*) as total_trades,
				SUM(CASE WHEN pn_l > 0 THEN 1 ELSE 0 END) as winning_trades,
				SUM(CASE WHEN pn_l < 0 THEN 1 ELSE 0 END) as losing_trades,
				ROUND(SUM(pn_l), 4) as total_pnl,
				ROUND(AVG(pn_l_percent), 2) as avg_pnl_percent,
				ROUND(SUM(value), 4) as total_volume,
				MIN(timestamp) as first_trade,
				MAX(timestamp) as last_trade
			FROM trades
			GROUP BY symbol
		`).Error; err != nil {
			log.Printf("Warning: Failed to create trade_summary view: %v", err)
		}
	}

	return db, nil
}

// backupDatabase creates a backup of the database file
func backupDatabase(dbPath string) error {
	// Create backups directory
	backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("trading_bot_%s.db", timestamp))

	// Copy database file
	source, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open source database: %v", err)
	}
	defer source.Close()

	destination, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("failed to copy database: %v", err)
	}

	// Clean up old backups (keep last 7 days)
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(backupDir, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}
		if time.Since(fileInfo.ModTime()) > 7*24*time.Hour {
			os.Remove(filePath)
		}
	}

	return nil
}

// createViews is now only used for maintenance/updates
func (db *Database) createViews() error {
	// Only try to update view if trades table exists
	if db.gorm.Migrator().HasTable("trades") {
		err := db.gorm.Exec(`
			CREATE VIEW IF NOT EXISTS trade_summary AS
			SELECT
				symbol,
				COUNT(*) as total_trades,
				SUM(CASE WHEN pn_l > 0 THEN 1 ELSE 0 END) as winning_trades,
				SUM(CASE WHEN pn_l < 0 THEN 1 ELSE 0 END) as losing_trades,
				ROUND(SUM(pn_l), 4) as total_pnl,
				ROUND(AVG(pn_l_percent), 2) as avg_pnl_percent,
				ROUND(SUM(value), 4) as total_volume,
				MIN(timestamp) as first_trade,
				MAX(timestamp) as last_trade
			FROM trades
			GROUP BY symbol
		`).Error

		if err != nil {
			return fmt.Errorf("error creating trade_summary view: %v", err)
		}
	}
	return nil
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
			SUM(CASE WHEN pn_l > 0 THEN 1 ELSE 0 END) as winning_trades,
			SUM(CASE WHEN pn_l < 0 THEN 1 ELSE 0 END) as losing_trades,
			ROUND(SUM(pn_l), 4) as total_pnl,
			ROUND(AVG(pn_l_percent), 2) as avg_pnl_percent,
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
