package strategy

import (
	"github.com/marwanbukhori/player-cryptobot/internal/models"
)

// Strategy interface defines the common behavior for all trading strategies
type Strategy interface {
	Analyze(data *models.MarketData) *models.Signal
}
