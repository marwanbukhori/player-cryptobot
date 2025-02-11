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

			// Analyze market data
			signal := strategy.Analyze(&models.MarketData{
				Symbol: pair,
				Price:  price,
				Time:   time.Now(),
			})

			if signal != nil {
				log.Info("🔍 %s Analysis - Price: %.2f USDT, Signal: %s",
					pair, price, signal.Action)

				// Get current position
				balances, err := exchange.GetBalance()
				if err != nil {
					log.Error("Error getting balance: %v", err)
					continue
				}

				btcBalance := balances["BTC"]

				// Buy signal handling
				if signal.Action == "BUY" {
					usdtBalance := balances["USDT"]
					if usdtBalance < cfg.MinOrderSize {
						log.Debug("💰 Insufficient USDT balance (%.2f) for trading", usdtBalance)
						continue
					}

					log.Info("🟢 BUY Signal - %s at %.2f USDT (Balance: %.2f USDT)",
						pair, price, usdtBalance)

					// Calculate position size based on available USDT and risk management
					stopLoss := price * 0.995 // 0.5% stop loss
					quantity, err := riskManager.CalculatePositionSize(price, stopLoss)
					if err != nil {
						log.Error("Error calculating position size: %v", err)
						continue
					}

					// Ensure we don't exceed available USDT and respect minimum order size
					maxQuantity := (usdtBalance * 0.95) / price
					if quantity > maxQuantity {
						quantity = maxQuantity
					}

					// Ensure minimum order size
					minOrderValue := quantity * price
					if minOrderValue < cfg.MinOrderSize {
						log.Debug("💡 Order value (%.2f) below minimum (%.2f), skipping", minOrderValue, cfg.MinOrderSize)
						continue
					}

					// Generate position ID for tracking
					positionID := generateUUID()

					order := &models.Order{
						Symbol:    signal.Symbol,
						Side:      "BUY",
						Type:      "MARKET",
						Quantity:  quantity,
						Price:     price,
						Timestamp: time.Now(),
					}

					// Place the buy order
					if err := exchange.PlaceOrder(order); err != nil {
						log.Error("❌ Failed to place BUY order: %v", err)
						notifier.NotifyError(err)
						continue
					}

					// Save trade to database
					trade := &models.Trade{
						Symbol:     order.Symbol,
						Side:       order.Side,
						Price:      price,
						Quantity:   order.Quantity,
						Value:      price * order.Quantity,
						Fee:        price * order.Quantity * 0.001,
						Timestamp:  order.Timestamp,
						PositionID: positionID,
						Status:     "OPEN",
					}

					if err := exchange.SaveTrade(trade); err != nil {
						log.Error("Error saving trade: %v", err)
						continue
					}

					// Notify about successful buy
					notifier.NotifyTrade(order.Symbol, order.Side, price, order.Quantity)

					log.Info("✅ BUY Order Filled - %s: %.8f at %.2f USDT (Total: %.2f USDT)",
						pair, quantity, price, quantity*price)
				}

				// Sell signal handling
				lastBuy, err := exchange.GetLastBuyTrade(signal.Symbol)
				if err != nil {
					if err.Error() != "no previous buy trade found" {
						log.Error("Error getting last buy trade: %v", err)
					}
					continue
				}

				if lastBuy != nil {
					potentialProfit := ((price - lastBuy.Price) / lastBuy.Price) * 100

					// Sell when:
					// 1. We get a SELL signal or meet profit target
					// 2. We have crypto balance to sell
					if (signal.Action == "SELL" || potentialProfit >= 2.0) && btcBalance > 0.0001 {
						log.Info("🔴 SELL Signal - %s at %.2f USDT (Entry: %.2f, PnL: %.2f%%)",
							pair, price, lastBuy.Price, potentialProfit)

						sellQuantity := btcBalance

						// Tiered exit system
						if potentialProfit >= 5.0 {
							sellQuantity = btcBalance * 0.5 // Sell 50% at 5% profit
							log.Info("📈 Taking 50% profit at %.2f%%", potentialProfit)
						} else if potentialProfit >= 3.0 {
							sellQuantity = btcBalance * 0.3 // Sell 30% at 3% profit
							log.Info("📈 Taking 30% profit at %.2f%%", potentialProfit)
						}

						order := &models.Order{
							Symbol:    signal.Symbol,
							Side:      "SELL",
							Type:      "MARKET",
							Quantity:  sellQuantity,
							Price:     price,
							Timestamp: time.Now(),
						}

						// Place the sell order
						if err := exchange.PlaceOrder(order); err != nil {
							log.Error("❌ Failed to place SELL order: %v", err)
							notifier.NotifyError(err)
							continue
						}

						// Save trade to database with proper position linking
						sellTrade := &models.Trade{
							Symbol:     order.Symbol,
							Side:       order.Side,
							Price:      price,
							Quantity:   order.Quantity,
							Value:      price * order.Quantity,
							Fee:        price * order.Quantity * 0.001,
							Timestamp:  order.Timestamp,
							PositionID: lastBuy.PositionID, // Link to the buy trade
							Status:     "CLOSED",
							PnL:        (price - lastBuy.Price) * order.Quantity,
							PnLPercent: potentialProfit,
						}

						if err := exchange.SaveTrade(sellTrade); err != nil {
							log.Error("Error saving trade: %v", err)
							continue
						}

						// Notify about successful sell
						notifier.NotifyTrade(order.Symbol, order.Side, price, order.Quantity)

						log.Info("✅ SELL Order Filled - %s: %.8f at %.2f USDT (PnL: %.2f%%)",
							pair, order.Quantity, price, potentialProfit)
					}
				}
			}
		}

		// Adjust trading frequency to prevent rapid trades
		time.Sleep(time.Minute) // Back to 1-minute interval
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
