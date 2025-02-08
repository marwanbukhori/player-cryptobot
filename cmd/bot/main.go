package main

import (
	"crypto/rand"
	"encoding/base32"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/marwanbukhori/player-cryptobot/internal/config"
	"github.com/marwanbukhori/player-cryptobot/internal/database"
	"github.com/marwanbukhori/player-cryptobot/internal/exchange"
	"github.com/marwanbukhori/player-cryptobot/internal/logger"
	"github.com/marwanbukhori/player-cryptobot/internal/models"
	"github.com/marwanbukhori/player-cryptobot/internal/notifications"
	"github.com/marwanbukhori/player-cryptobot/internal/risk"
	"github.com/marwanbukhori/player-cryptobot/internal/strategy"
	"github.com/marwanbukhori/player-cryptobot/internal/web"
)

func main() {
	// Initialize logger first
	log := logger.NewLogger()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("Failed to load config: %v", err)
		os.Exit(1)
	}
	log.Info("Configuration loaded successfully")

	// Initialize database
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		log.Error("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	log.Info("Database initialized successfully")

	// Initialize exchange with the database instance
	exchange, err := exchange.NewExchange(cfg, db)
	if err != nil {
		log.Error("Failed to initialize exchange: %v", err)
		os.Exit(1)
	}
	log.Info("Connected to Binance successfully")

	// Initialize strategy
	strategy := strategy.NewMeanReversionStrategy()

	// Initialize risk manager
	riskManager := risk.NewRiskManager(
		cfg.InitialInvestment,
		cfg.MaxDrawdown,
		cfg.RiskPerTrade,
		cfg.AggressiveFactor,
		cfg.EnableCompounding,
	)

	// Initialize telegram notifier
	notifier := notifications.NewTelegramNotifier(
		cfg.TelegramToken,
		cfg.TelegramChatID,
	)

	// Start web dashboard
	go func() {
		server := web.NewServer(exchange, ":8080")
		if err := server.Start(); err != nil {
			log.Error("Failed to start web server: %v", err)
		}
	}()

	// Start trading loop
	for {
		for _, pair := range cfg.TradingPairs {
			price, err := exchange.GetPrice(pair)
			if err != nil {
				log.Error("Error getting price for %s: %v", pair, err)
				notifier.NotifyError(err)
				continue
			}

			log.Debug("Current price for %s: %.2f", pair, price)

			// Analyze market data
			signal := strategy.Analyze(&models.MarketData{
				Symbol: pair,
				Price:  price,
				Time:   time.Now(),
			})

			if signal != nil {
				// Get current position
				balances, err := exchange.GetBalance()
				if err != nil {
					log.Error("Error getting balance: %v", err)
					continue
				}

				btcBalance := balances["BTC"]

				// Buy when:
				// 1. We get a BUY signal from strategy
				// 2. We have enough USDT balance
				if signal.Action == "BUY" {
					// Get USDT balance instead of BTC
					usdtBalance := balances["USDT"]
					if usdtBalance < cfg.MinOrderSize {
						log.Debug("Insufficient USDT balance for trading")
						continue
					}

					log.Info("Placing BUY order - Price: %.2f, Available USDT: %.2f",
						price, usdtBalance)

					// Calculate position size based on available USDT and risk management
					stopLoss := price * 0.995 // 0.5% stop loss
					quantity, err := riskManager.CalculatePositionSize(price, stopLoss)
					if err != nil {
						log.Error("Error calculating position size: %v", err)
						continue
					}

					// Ensure we don't exceed available USDT
					maxQuantity := (usdtBalance * 0.95) / price
					if quantity > maxQuantity {
						quantity = maxQuantity
					}

					order := &models.Order{
						Symbol:    signal.Symbol,
						Side:      "BUY",
						Type:      "MARKET",
						Quantity:  quantity,
						Price:     price,
						Timestamp: time.Now(),
					}

					if err := exchange.PlaceOrder(order); err != nil {
						log.Error("Error placing order: %v", err)
						notifier.NotifyError(err)
						continue
					}

					// Save trade to database
					trade := &models.Trade{
						Symbol:    order.Symbol,
						Side:      order.Side,
						Price:     price, // Use the actual price from earlier
						Quantity:  order.Quantity,
						Value:     price * order.Quantity, // Calculate using actual price
						Fee:       price * order.Quantity * 0.001,
						Timestamp: order.Timestamp,
					}

					if err := exchange.SaveTrade(trade); err != nil {
						log.Error("Error saving trade: %v", err)
						continue
					}

					// Notify with proper price formatting
					notifier.NotifyTrade(order.Symbol, order.Side, price, order.Quantity) // Use actual price

					// When placing buy order
					positionID := generateUUID()
					trade.PositionID = positionID
					trade.Status = "OPEN"
				}

				// Before the sell condition, add:
				lastBuy, err := exchange.GetLastBuyTrade(signal.Symbol)
				if err != nil {
					log.Error("Error getting last buy trade: %v", err)
					continue
				}
				potentialProfit := ((price - lastBuy.Price) / lastBuy.Price) * 100

				// Sell when:
				// 1. We get a SELL signal or meet profit target
				// 2. We have crypto balance to sell
				if (signal.Action == "SELL" || potentialProfit >= 2.0) && btcBalance > 0.0001 {
					log.Info("Placing SELL order - Entry: %.2f, Current: %.2f, Profit: %.2f%%",
						lastBuy.Price, price, potentialProfit)

					order := &models.Order{
						Symbol:    signal.Symbol,
						Side:      "SELL",
						Type:      "MARKET",
						Quantity:  btcBalance,
						Price:     price,
						Timestamp: time.Now(),
					}

					// Tiered exit system
					if potentialProfit >= 5.0 {
						order.Quantity = btcBalance * 0.5 // Sell 50% at 5% profit
					} else if potentialProfit >= 3.0 {
						order.Quantity = btcBalance * 0.3 // Sell 30% at 3% profit
					}

					if err := exchange.PlaceOrder(order); err != nil {
						log.Error("Error placing order: %v", err)
						notifier.NotifyError(err)
						continue
					}

					// When placing sell order
					sellTrade := &models.Trade{
						Symbol:    order.Symbol,
						Side:      order.Side,
						Price:     price,
						Quantity:  order.Quantity,
						Value:     price * order.Quantity,
						Fee:       price * order.Quantity * 0.001,
						Timestamp: order.Timestamp,
					}

					if err := exchange.SaveTrade(sellTrade); err != nil {
						log.Error("Error saving trade: %v", err)
						continue
					}

					sellTrade.PositionID = lastBuy.PositionID
					sellTrade.Status = "CLOSED"
				}
			}
		}

		time.Sleep(time.Minute) // Adjust frequency as needed
	}
}

func printTradingSummary(exchange exchange.Exchange, log *logger.Logger) {
	summary, err := exchange.GetTradingSummary()
	if err != nil {
		log.Error("Failed to get trading summary: %v", err)
		return
	}

	log.Info("=== Trading Summary ===")
	log.Info("Total Trades: %d", len(summary))

	var totalPnL float64
	for _, s := range summary {
		totalPnL += s.TotalPnL
	}

	log.Info("Total P&L: $%.2f", totalPnL)
}

func CalculateRealizedPnl(exchange exchange.Exchange) (map[string]float64, error) {
	trades, err := exchange.GetAllTrades()
	if err != nil {
		return nil, err
	}

	pnl := make(map[string]float64)
	for _, trade := range trades {
		if trade.Side == "SELL" {
			pnl[trade.PositionID] = trade.PnL
		}
	}
	return pnl, nil
}

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic("failed to generate UUID: " + err.Error())
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}
