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

/*
	BinanceExchange

*  implements the Exchange interface for Binance
*/
type binanceExchange struct {
	client *binance.Client
	config *config.Config
	db     *database.Database
}

/*
	NewExchange

*  create a new exchange instance
*/
func NewExchange(config *config.Config, db *database.Database) (Exchange, error) {
	var log = logrus.New()

	/* Get current IP
	*  to check if bot running from whitelist IP
	 */
	ip, err := getPublicIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get IP: %v", err)
	}
	log.Infof("Bot running from IP: %s", ip)

	/*
		Create a new Binance client
		*  using the API key and secret
	*/
	client := binance.NewClient(config.BINANCE_API_KEY, config.BINANCE_API_SECRET)

	/*
		Set a longer recvWindow (default is 5000ms)
		* What is recvWindow?
		*  it is the time in milliseconds
		*  that the client will wait for the server to respond
		*  to avoid rate limit errors
	*/
	client.TimeOffset = 0
	client.BaseURL = "https://api.binance.com"
	client.UserAgent = "Mozilla/5.0"

	/*
		First, get server time and calculate offset
		*  to get the time offset between the client and the server
	*/
	serverTime, err := client.NewServerTimeService().Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get server time: %v", err)
	}

	/*
		Calculate time offset
		*  to get the time offset between the client and the server
	*/
	timeOffset := serverTime - time.Now().UnixMilli()
	client.TimeOffset = timeOffset

	log.Infof("Time offset with Binance: %dms", timeOffset)

	/*
		Test with simple ping
		*  to check if the client is working
	*/
	err = client.NewPingService().Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping Binance: %v", err)
	}
	log.Info("Ping successful")

	/*
		Now try account info with retries
		*  to check if the account info is working
	*/
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

	/*
		Log non-zero balances
		*  to check if the balances are working
	*/
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

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func roundToValidQuantity(quantity float64) float64 {
	/* Binance minimum quantities */
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

/**
*
* All the exchange Query Functions
* TODO: To verify each query
* - PlaceOrder
* - GetPrice
* - placeStopLossOrder
* - GetHistoricalData
* - GetBalance
* - GetTradingSummary
* - GetLastBuyTrade
* - GetAllTrades
* - GetRecentTrades
* - SaveTrade
* - GetOpenPositions
* - GetOpenPosition
**/

/*
	PlaceOrder

*  place an order on the exchange
*/
func (b *binanceExchange) PlaceOrder(order *models.Order) error {

	/* Get current price for accurate calculations
	 */
	currentPrice, err := b.GetPrice(order.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get current price: %v", err)
	}
	order.Price = currentPrice // Use actual current price

	/* Check stop loss first before placing new orders
	 */
	if order.Side == "SELL" {
		if order.Type == "MARKET" && currentPrice < order.StopLossPrice {
			return fmt.Errorf("market sell blocked: price %.2f < stop loss %.2f", currentPrice, order.StopLossPrice)
		}
	}

	/* Check balance before trading
	 */
	balances, err := b.GetBalance()
	if err != nil {
		return fmt.Errorf("failed to get balance: %v", err)
	}

	if order.Side == "BUY" {
		/* Check USDT balance for buying */
		if usdtBalance := balances["USDT"]; usdtBalance < (order.Price * order.Quantity) {
			return fmt.Errorf("insufficient USDT balance: have %.2f, need %.2f",
				usdtBalance, order.Price*order.Quantity)
		}
	} else {
		/* Check BTC balance for selling */
		if btcBalance := balances["BTC"]; btcBalance < order.Quantity {
			return fmt.Errorf("insufficient BTC balance: have %.8f, need %.8f",
				btcBalance, order.Quantity)
		}
	}

	/* Round quantity to valid lot size
	 */
	quantity := roundToValidQuantity(order.Quantity)
	if quantity <= 0 {
		return fmt.Errorf("order quantity too small: %.8f", order.Quantity)
	}

	/* Place the actual order
	 */
	orderService := b.client.NewCreateOrderService().
		Symbol(order.Symbol).
		Side(binance.SideType(order.Side)).
		Type(binance.OrderTypeMarket).
		Quantity(fmt.Sprintf("%.8f", quantity)).
		NewOrderRespType("FULL")

	/* Execute spot order
	 */
	result, err := orderService.Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to place spot order: %v", err)
	}

	/* Update order details with actual execution price and quantity
	 */
	order.Price, _ = strconv.ParseFloat(result.Price, 64)
	order.Quantity, _ = strconv.ParseFloat(result.ExecutedQuantity, 64)

	/* Immediately place stop loss order after successful buy
	 */
	if order.Side == "BUY" {
		stopLossPrice := order.Price * 0.995
		if stopLossPrice <= 0 {
			return fmt.Errorf("invalid stop loss price: %.2f", stopLossPrice)
		}
		limitPrice := stopLossPrice * 0.998
		stopLossOrder := &models.Order{
			Symbol:        order.Symbol,
			Side:          "SELL",
			Type:          "STOP_LOSS_LIMIT",
			Quantity:      order.Quantity,
			Price:         limitPrice,
			StopLossPrice: stopLossPrice,
			Timestamp:     time.Now(),
		}

		/* Place stop loss order
		 */
		if err := b.placeStopLossOrder(stopLossOrder); err != nil {
			return fmt.Errorf("failed to place stop loss: %v", err)
		}
	}

	return nil
}

/*
	GetPrice

*  get the current price of the symbol
*/
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

/*
	placeStopLossOrder

*  place a stop loss order on the exchange
*/
func (b *binanceExchange) placeStopLossOrder(order *models.Order) error {
	orderService := b.client.NewCreateOrderService().
		Symbol(order.Symbol).
		Side(binance.SideType(order.Side)).
		Type(binance.OrderTypeStopLossLimit).
		TimeInForce(binance.TimeInForceTypeGTC).
		Quantity(fmt.Sprintf("%.8f", order.Quantity)).
		Price(fmt.Sprintf("%.2f", order.Price)).
		StopPrice(fmt.Sprintf("%.2f", order.StopLossPrice))

	_, err := orderService.Do(context.Background())
	return err
}

/*
	GetHistoricalData

*  get the historical data of the symbol
*/
func (b *binanceExchange) GetHistoricalData(symbol string, interval string, limit int) ([]models.Kline, error) {
	/* Example: interval = "1m", "5m", "1h", "1d" */
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

/*
	GetBalance

*  get the balance of the account
*/
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

/*
	GetTradingSummary

*  get the trading summary of the account
*/
func (b *binanceExchange) GetTradingSummary() ([]models.TradingSummary, error) {
	return b.db.GetTradingSummary()
}

/*
	GetLastBuyTrade

*  get the last buy trade of the symbol
*/
func (b *binanceExchange) GetLastBuyTrade(symbol string) (*models.Trade, error) {
	return b.db.GetLastBuyTrade(symbol)
}

/*
	GetAllTrades

*  get all the trades of the account
*/
func (b *binanceExchange) GetAllTrades() ([]models.Trade, error) {
	return b.db.GetRecentTrades(1000) // Limit to last 1000 trades for performance
}

/*
	GetRecentTrades

*  get the recent trades of the account
*/
func (b *binanceExchange) GetRecentTrades(limit int) ([]models.Trade, error) {
	return b.db.GetRecentTrades(limit)
}

/*
	SaveTrade

*  save the trade to the database
*/
func (b *binanceExchange) SaveTrade(trade *models.Trade) error {
	return b.db.SaveTrade(trade)
}

/*
	GetOpenPositions

*  get the open positions of the account
*/
func (b *binanceExchange) GetOpenPositions() ([]models.Trade, error) {
	return b.db.GetOpenPositions()
}

/*
	GetOpenPosition

*  get the last buy trade of the symbol
*/
func (b *binanceExchange) GetOpenPosition(symbol string) (*models.Trade, error) {
	trades, err := b.GetTrades(symbol)
	if err != nil {
		return nil, err
	}

	// Find latest BUY without corresponding SELL
	for i := len(trades) - 1; i >= 0; i-- {
		if trades[i].Side == "BUY" && trades[i].Status == "OPEN" {
			return trades[i], nil
		}
	}

	return nil, nil
}

/*
	GetTrades

*  get the trades of the symbol
*/
func (b *binanceExchange) GetTrades(symbol string) ([]*models.Trade, error) {
	// Get all trades from database
	allTrades, err := b.db.GetRecentTrades(1000)
	if err != nil {
		return nil, err
	}

	// Filter by symbol
	var filtered []*models.Trade
	for i := range allTrades {
		if allTrades[i].Symbol == symbol {
			filtered = append(filtered, &allTrades[i])
		}
	}
	return filtered, nil
}

/*
	UpdateTradeStatus

*  update the status of the trade
*/
func (b *binanceExchange) UpdateTradeStatus(positionID string, status string) error {
	return b.db.UpdateTradeStatus(positionID, status)
}
