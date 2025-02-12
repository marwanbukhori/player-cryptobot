package exchange

import (
	"github.com/marwanbukhori/player-cryptobot/internal/models"
)

/*
*  Exchange defines the interface for cryptocurrency exchange operations
 */
type Exchange interface {
	GetPrice(symbol string) (float64, error)
	PlaceOrder(order *models.Order) error
	GetBalance() (map[string]float64, error)
	GetHistoricalData(symbol string, interval string, limit int) ([]models.Kline, error)
	GetTradingSummary() ([]models.TradingSummary, error)
	GetLastBuyTrade(symbol string) (*models.Trade, error)
	GetAllTrades() ([]models.Trade, error)
	GetRecentTrades(limit int) ([]models.Trade, error)
	SaveTrade(trade *models.Trade) error
	GetOpenPositions() ([]models.Trade, error)
	GetOpenPosition(symbol string) (*models.Trade, error)
	GetTrades(symbol string) ([]*models.Trade, error)
	UpdateTradeStatus(positionID string, status string) error
}
