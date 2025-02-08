package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/marwanbukhori/player-cryptobot/internal/config"
	"github.com/marwanbukhori/player-cryptobot/internal/database"
	"github.com/marwanbukhori/player-cryptobot/internal/models"
	"github.com/sirupsen/logrus"
)

// BinanceExchange implements the Exchange interface for Binance
type binanceExchange struct {
	client *binance.Client
	config *config.Config
	db     *database.Database
}

// NewExchange creates a new exchange instance
func NewExchange(config *config.Config, db *database.Database) (Exchange, error) {
	var log = logrus.New()

	// Get current IP
	ip, err := getPublicIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get IP: %v", err)
	}
	log.Infof("Bot running from IP: %s", ip)

	client := binance.NewClient(config.APIKey, config.APISecret)

	// Set a longer recvWindow (default is 5000ms)
	client.TimeOffset = 0
	client.BaseURL = "https://api.binance.com"
	client.UserAgent = "Mozilla/5.0"

	// First, get server time and calculate offset
	serverTime, err := client.NewServerTimeService().Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get server time: %v", err)
	}

	// Calculate time offset
	timeOffset := serverTime - time.Now().UnixMilli()
	client.TimeOffset = timeOffset

	log.Infof("Time offset with Binance: %dms", timeOffset)

	// Test with simple ping
	err = client.NewPingService().Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping Binance: %v", err)
	}
	log.Info("Ping successful")

	// Now try account info with retries
	var account *binance.Account
	maxAttempts := 10 // Maximum number of attempts
	attempt := 1

	for {
		account, err = client.NewGetAccountService().Do(context.Background())
		if err == nil {
			log.Infof("Account access successful after %d attempts", attempt)
			break
		}

		if attempt >= maxAttempts {
			return nil, fmt.Errorf("failed to get account info after %d attempts: %v", maxAttempts, err)
		}

		backoff := time.Duration(attempt) * time.Second // Increasing backoff
		log.Warnf("Attempt %d: Failed to get account info: %v. Retrying in %v...", attempt, err, backoff)
		time.Sleep(backoff)
		attempt++
	}

	// Log non-zero balances
	for _, balance := range account.Balances {
		free, _ := strconv.ParseFloat(balance.Free, 64)
		locked, _ := strconv.ParseFloat(balance.Locked, 64)
		if free > 0 || locked > 0 {
			log.Infof("Balance %s: Free %.8f, Locked %.8f",
				balance.Asset, free, locked)
		}
	}

	exchange := &binanceExchange{
		client: client,
		config: config,
		db:     db,
	}

	return exchange, nil
}

func getPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}

func (b *binanceExchange) GetPrice(symbol string) (float64, error) {
	prices, err := b.client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	if len(prices) == 0 {
		return 0, fmt.Errorf("no price found for symbol %s", symbol)
	}
	return strconv.ParseFloat(prices[0].Price, 64)
}

func (b *binanceExchange) PlaceOrder(order *models.Order) error {
	var log = logrus.New()

	// Get current price for accurate calculations
	currentPrice, err := b.GetPrice(order.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get current price: %v", err)
	}
	order.Price = currentPrice // Use actual current price

	// Calculate value before placing order
	value := order.Price * order.Quantity

	// Create trade record
	trade := &models.Trade{
		Symbol:    order.Symbol,
		Side:      order.Side,
		Price:     order.Price,
		Quantity:  order.Quantity,
		Value:     value,
		Fee:       value * 0.001, // 0.1% fee
		Timestamp: time.Now(),
	}

	// For SELL orders, calculate P&L
	if order.Side == "SELL" {
		lastBuy, err := b.GetLastBuyTrade(order.Symbol)
		if err == nil && lastBuy != nil {
			// Calculate actual profit/loss
			buyValue := lastBuy.Price * order.Quantity            // What we paid for current quantity
			sellValue := order.Price * order.Quantity             // What we got from selling
			totalFees := (buyValue * 0.001) + (sellValue * 0.001) // Both buy and sell fees

			trade.PnL = sellValue - buyValue - totalFees
			trade.PnLPercent = (trade.PnL / buyValue) * 100

			log.Infof("Trade details - Buy: %.2f, Sell: %.2f, Fees: %.2f, P&L: %.2f USDT (%.2f%%)",
				buyValue, sellValue, totalFees, trade.PnL, trade.PnLPercent)
		}
	}

	// Place the actual order
	// Check balance before trading
	balances, err := b.GetBalance()
	if err != nil {
		return fmt.Errorf("failed to get balance: %v", err)
	}

	if order.Side == "BUY" {
		// Check USDT balance for buying
		if usdtBalance := balances["USDT"]; usdtBalance < (order.Price * order.Quantity) {
			return fmt.Errorf("insufficient USDT balance: have %.2f, need %.2f",
				usdtBalance, order.Price*order.Quantity)
		}
	} else {
		// Check BTC balance for selling
		if btcBalance := balances["BTC"]; btcBalance < order.Quantity {
			return fmt.Errorf("insufficient BTC balance: have %.8f, need %.8f",
				btcBalance, order.Quantity)
		}
	}

	// Round quantity to valid lot size
	quantity := roundToValidQuantity(order.Quantity)
	if quantity <= 0 {
		return fmt.Errorf("order quantity too small: %.8f", order.Quantity)
	}

	orderService := b.client.NewCreateOrderService().
		Symbol(order.Symbol).
		Side(binance.SideType(order.Side)).
		Type(binance.OrderTypeMarket).
		Quantity(fmt.Sprintf("%.8f", quantity)).
		NewOrderRespType("FULL")

	// Execute spot order
	result, err := orderService.Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to place spot order: %v", err)
	}

	// Update order details
	order.Price, _ = strconv.ParseFloat(result.Price, 64)
	order.Quantity, _ = strconv.ParseFloat(result.ExecutedQuantity, 64)

	// Save trade to database
	if err := b.SaveTrade(trade); err != nil {
		log.Errorf("Failed to save trade: %v", err)
	}

	return nil
}

func (b *binanceExchange) GetHistoricalData(symbol string, interval string, limit int) ([]models.Kline, error) {
	// Example: interval = "1m", "5m", "1h", "1d"
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		symbol, interval, limit)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rawKlines [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawKlines); err != nil {
		return nil, err
	}

	klines := make([]models.Kline, len(rawKlines))
	for i, raw := range rawKlines {
		klines[i] = models.Kline{
			OpenTime:  int64(raw[0].(float64)),
			Open:      parseFloat(raw[1].(string)),
			High:      parseFloat(raw[2].(string)),
			Low:       parseFloat(raw[3].(string)),
			Close:     parseFloat(raw[4].(string)),
			Volume:    parseFloat(raw[5].(string)),
			CloseTime: int64(raw[6].(float64)),
		}
	}

	return klines, nil
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func (b *binanceExchange) GetBalance() (map[string]float64, error) {
	account, err := b.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %v", err)
	}

	balances := make(map[string]float64)
	for _, b := range account.Balances {
		if free, err := strconv.ParseFloat(b.Free, 64); err == nil {
			if free > 0 {
				balances[b.Asset] = free
			}
		}
	}
	return balances, nil
}

func (b *binanceExchange) GetTradingSummary() ([]models.TradingSummary, error) {
	return b.db.GetTradingSummary()
}

func (b *binanceExchange) GetLastBuyTrade(symbol string) (*models.Trade, error) {
	return b.db.GetLastBuyTrade(symbol)
}

func (b *binanceExchange) GetAllTrades() ([]models.Trade, error) {
	return b.db.GetRecentTrades(1000) // Limit to last 1000 trades for performance
}

func roundToValidQuantity(quantity float64) float64 {
	// Binance minimum quantities
	switch {
	case quantity < 0.00001:
		return 0
	case quantity < 0.001:
		return math.Floor(quantity*100000) / 100000 // 5 decimal places
	case quantity < 0.1:
		return math.Floor(quantity*1000) / 1000 // 3 decimal places
	default:
		return math.Floor(quantity*100) / 100 // 2 decimal places
	}
}

func (b *binanceExchange) GetRecentTrades(limit int) ([]models.Trade, error) {
	return b.db.GetRecentTrades(limit)
}

func (b *binanceExchange) SaveTrade(trade *models.Trade) error {
	return b.db.SaveTrade(trade)
}
