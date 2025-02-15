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
	/*
	* Initialize logger first
	 */
	log := logger.NewLogger()

	/*
	* Load configuration
	 */
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("Failed to load config: %v", err)
		os.Exit(1)
	}
	log.Info("Configuration loaded successfully")

	/*
	* Initialize database
	 */
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		log.Error("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	log.Info("Database initialized successfully")

	/*
	* Initialize exchange with the database instance
	 */
	exchange, err := exchange.NewExchange(cfg, db)
	if err != nil {
		log.Error("Failed to initialize exchange: %v", err)
		os.Exit(1)
	}
	log.Info("Connected to Binance successfully")

	/* Initialize strategy
	*  Currently using Mean Reversion Strategy
	*  Can add more in the future
	 */
	strategy := strategy.NewMeanReversionStrategy()

	/*
	* Initialize risk manager
	 */
	riskManager := risk.NewRiskManager(
		cfg.InitialInvestment,
		cfg.MaxDrawdown,
		cfg.RiskPerTrade,
		cfg.AggressiveFactor,
		cfg.EnableCompounding,
	)

	/*
	* Initialize telegram notifier
	 */
	notifier := notifications.NewTelegramNotifier(
		cfg.TelegramToken,
		cfg.TelegramChatID,
	)

	/*
	* Start web dashboard
	 */
	go func() {
		server := web.NewServer(exchange, ":8080")
		if err := server.Start(); err != nil {
			log.Error("Failed to start web server: %v", err)
		}
	}()

	/* Start Trading Loop

	* What does this do?
	* - Get the price of the trading pair
	* - Get the last buy trade
	* - Get the open position
	* - Calculate the potential profit
	* - If the potential profit is less than -5%, set the signal to SELL
	* - If the signal is SELL, sell the position
	* - If the signal is BUY, buy the position
	 */
	for {
		for _, pair := range cfg.TradingPairs {
			price, err := exchange.GetPrice(pair)
			if err != nil {
				log.Error("Error getting price for %s: %v", pair, err)
				notifier.NotifyError(err)
				continue
			}

			/* What is a signal?
			* It is a signal that the trading bot will follow
			* which based on the strategy
			 */
			var signal *models.Signal

			/* Get the last buy trade there is no error,
			calculate the potential profit
			*/
			lastBuy, err := exchange.GetOpenPosition(pair)
			if err == nil && lastBuy != nil {

				/* TODO: Profit Calculation might need to be in a different function
				to handle more complex calculations
				*/
				potentialProfit := ((price - lastBuy.Price) / lastBuy.Price) * 100
				log.Info("📊 %s Current Price: %.2f | Entry: %.2f | PnL: %.2f%%",
					pair, price, lastBuy.Price, potentialProfit)

				/*
				* Emergency sell check
				* if the potential profit is less than -5%, set the signal to SELL
				 */
				if potentialProfit < -5.0 {
					log.Error("⚠️🔴 Emergency sell at 5%% loss")
					signal = &models.Signal{
						Symbol: pair,
						Action: "SELL",
					}
				}
			}

			/*
			* Analyze market data
			 */
			strategySignal := strategy.Analyze(&models.MarketData{
				Symbol: pair,
				Price:  price,
				Time:   time.Now(),
			})

			/* if the strategy signal is not nil, set the signal to the strategy signal
			* and log the signal
			 */
			if strategySignal != nil {
				signal = strategySignal
				log.Info("🔍 %s Analysis - Price: %.2f USDT, Signal: %s",
					pair, price, signal.Action)

				/* Get current account balance */
				balances, err := exchange.GetBalance()
				if err != nil {
					log.Error("Error getting balance: %v", err)
					continue
				}

				/* Specify what balance */
				btcBalance := balances["BTC"]

				/*
				* BUY signal handling
				 */
				if signal.Action == "BUY" {
					usdtBalance := balances["USDT"]
					if usdtBalance < cfg.MinOrderSize {
						log.Debug("💰 Insufficient USDT balance (%.2f) for trading", usdtBalance)
						continue
					}

					log.Info("🟢 BUY Signal - %s at %.2f USDT (Balance: %.2f USDT)",
						pair, price, usdtBalance)

					/* Calculate position size based on available USDT and risk management */
					stopLoss := price * 0.995 // 0.5% stop loss
					quantity, err := riskManager.CalculatePositionSize(price, stopLoss)
					if err != nil {
						log.Error("Error calculating position size: %v", err)
						continue
					}

					/* Ensure we don't exceed available USDT and respect minimum order size */
					maxQuantity := (usdtBalance * 0.95) / price
					if quantity > maxQuantity {
						quantity = maxQuantity
					}

					/* Ensure minimum order size */
					minOrderValue := quantity * price
					if minOrderValue < cfg.MinOrderSize {
						log.Debug("💡 Order value (%.2f) below minimum (%.2f), skipping", minOrderValue, cfg.MinOrderSize)
						continue
					}

					/* Generate position ID for tracking */
					positionID := generateUUID()

					order := &models.Order{
						Symbol:    signal.Symbol,
						Side:      "BUY",
						Type:      "MARKET",
						Quantity:  quantity,
						Price:     price,
						Timestamp: time.Now(),
					}

					/* Place the buy order */
					if err := exchange.PlaceOrder(order); err != nil {
						log.Error("❌ Failed to place BUY order: %v", err)
						notifier.NotifyError(err)
						continue
					}

					/* Save trade to database */
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

					/* Notify about successful buy */
					notifier.NotifyTrade(order.Symbol, order.Side, price, order.Quantity)

					log.Info("✅ BUY Order Filled - %s: %.8f at %.2f USDT (Total: %.2f USDT)",
						pair, quantity, price, quantity*price)
				}

				/*
				* SELL signal handling
				 */

				/*
				* If the last buy trade is not nil, calculate the potential profit
				* and check if the potential profit is less than -8%, set the signal to SELL
				 */
				if lastBuy != nil {
					potentialProfit := ((price - lastBuy.Price) / lastBuy.Price) * 100

					/* Added protection to sell the position if the potential profit is less than -8% */
					if potentialProfit < -8.0 {
						log.Error("⚠️🔴 Emergency sell at 8%% loss")
						signal.Action = "SELL"
					}

					/*
					* Sell when:
					* 1. We get a SELL signal or meet profit target
					* 2. We have crypto balance to sell
					 */
					if (signal.Action == "SELL" && btcBalance > 0.0001 && potentialProfit >= 0) || potentialProfit >= 2.0 {
						log.Info("🔴 SELL Signal - %s at %.2f USDT (Entry: %.2f, PnL: %.2f%%)",
							pair, price, lastBuy.Price, potentialProfit)

						sellQuantity := btcBalance

						/*
						* Tiered exit system
						* TODO: To verify is this relevant?

						*  - Sell 50% at 5% profit
						*  - Sell 30% at 3% profit
						 */
						if potentialProfit >= 5.0 {
							sellQuantity = btcBalance * 0.5 // Sell 50% at 5% profit
							log.Info("📈 Taking 50%% profit at %.2f%%", potentialProfit)
						} else if potentialProfit >= 3.0 {
							sellQuantity = btcBalance * 0.3 // Sell 30% at 3% profit
							log.Info("📈 Taking 30%% profit at %.2f%%", potentialProfit)
						}

						order := &models.Order{
							Symbol:    signal.Symbol,
							Side:      "SELL",
							Type:      "MARKET",
							Quantity:  sellQuantity,
							Price:     price,
							Timestamp: time.Now(),
						}

						/* Place the sell order */
						if err := exchange.PlaceOrder(order); err != nil {
							log.Error("❌ Failed to place SELL order: %v", err)
							notifier.NotifyError(err)
							continue
						}

						/* Before saving the sell trade, update the original BUY trade status */
						if err := exchange.UpdateTradeStatus(lastBuy.PositionID, "CLOSED"); err != nil {
							log.Error("Error closing position: %v", err)
						}

						/* Save trade to database with proper position linking */
						sellTrade := &models.Trade{
							Symbol:     order.Symbol,
							Side:       order.Side,
							Price:      price,
							Quantity:   order.Quantity,
							Value:      price * order.Quantity,
							Fee:        price * order.Quantity * 0.001,
							Timestamp:  order.Timestamp,
							PositionID: lastBuy.PositionID,
							Status:     "CLOSED",
							PnL:        (price - lastBuy.Price) * order.Quantity,
							PnLPercent: potentialProfit,
						}

						if err := exchange.SaveTrade(sellTrade); err != nil {
							log.Error("Error saving trade: %v", err)
							continue
						}

						/* Notify about successful sell */
						notifier.NotifyTrade(order.Symbol, order.Side, price, order.Quantity)

						log.Info("✅ SELL Order Filled - %s: %.8f at %.2f USDT (PnL: %.2f%%)",
							pair, order.Quantity, price, potentialProfit)
					}
				}
			}
		}

		/* Adjust trading frequency to prevent rapid trades
		* Currently: 10 seconds
		 */
		time.Sleep(10 * time.Second)
	}
}

/*
*  TODO: Verify if this is relevant
*  Print the trading summary
 */
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

/*
*  TODO: Verify if this legit
*  Calculate the realized PnL
 */
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

/*
*  Generate a UUID
 */
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic("failed to generate UUID: " + err.Error())
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}
