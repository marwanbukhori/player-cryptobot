package backtest

import (
	"time"

	"github.com/marwanbukhori/player-cryptobot/internal/models"
	"github.com/marwanbukhori/player-cryptobot/internal/strategy"
)

type MarketData struct {
	Time   time.Time
	Price  float64
	Volume float64
}

type BacktestResult struct {
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	TotalProfit   float64
	MaxDrawdown   float64
	WinRate       float64
	Trades        []models.Order
}

type Backtester struct {
	strategy strategy.Strategy
	data     []MarketData
	balance  float64
}

func NewBacktester(strategy strategy.Strategy, initialBalance float64) *Backtester {
	return &Backtester{
		strategy: strategy,
		balance:  initialBalance,
	}
}

func (b *Backtester) Run() BacktestResult {
	var result BacktestResult
	var position *models.Order
	maxBalance := b.balance
	minDrawdown := 0.0

	for i, data := range b.data {
		signal := b.strategy.Analyze(&models.MarketData{
			Symbol: "BTCUSDT",
			Price:  data.Price,
			Time:   data.Time,
		})

		if signal != nil {
			if position == nil && signal.Action == "BUY" {
				// Open position
				position = &models.Order{
					Symbol:    "BTCUSDT",
					Side:      "BUY",
					Price:     data.Price,
					Quantity:  b.balance / data.Price,
					Timestamp: data.Time,
				}
				result.TotalTrades++
			} else if position != nil && signal.Action == "SELL" {
				// Close position
				profit := (data.Price - position.Price) * position.Quantity
				b.balance += profit

				if profit > 0 {
					result.WinningTrades++
				} else {
					result.LosingTrades++
				}

				result.TotalProfit += profit
				result.Trades = append(result.Trades, *position)

				// Update drawdown
				if b.balance > maxBalance {
					maxBalance = b.balance
				}
				currentDrawdown := (maxBalance - b.balance) / maxBalance
				if currentDrawdown > minDrawdown {
					minDrawdown = currentDrawdown
				}

				position = nil
			}
		}

		if i == len(b.data)-1 {
			result.MaxDrawdown = minDrawdown
			result.WinRate = float64(result.WinningTrades) / float64(result.TotalTrades)
		}
	}

	return result
}

func (b *Backtester) LoadData(data []MarketData) {
	b.data = data
}

type Result struct {
	TotalTrades int
	WinRate     float64
	ProfitLoss  float64
}

func Run(data []models.Kline, strategy *strategy.MeanReversionStrategy, initialBalance float64) Result {
	var result Result
	balance := initialBalance
	position := 0.0

	for _, candle := range data {
		signal := strategy.Analyze(&models.MarketData{
			Symbol: "BTCUSDT",
			Price:  candle.Close,
			Time:   time.Unix(candle.CloseTime/1000, 0),
		})

		if signal != nil {
			result.TotalTrades++
			if signal.Action == "BUY" && position == 0 {
				position = balance / candle.Close
				balance = 0
			} else if signal.Action == "SELL" && position > 0 {
				balance = position * candle.Close
				if balance > initialBalance {
					result.WinRate++
				}
				position = 0
			}
		}
	}

	result.ProfitLoss = balance + (position * data[len(data)-1].Close) - initialBalance
	if result.TotalTrades > 0 {
		result.WinRate = (result.WinRate / float64(result.TotalTrades)) * 100
	}

	return result
}
