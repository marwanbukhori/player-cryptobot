# Exchange Module Documentation

## Overview

The exchange module provides a clean interface for interacting with cryptocurrency exchanges. Currently, it implements support for the Binance exchange with a flexible design that allows adding more exchanges in the future.

## Components

### Exchange Interface

The core interface that defines exchange operations:

```go
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
}
```

### Binance Implementation

The Binance exchange implementation of the Exchange interface:

```go
type binanceExchange struct {
    client *binance.Client
    config *config.Config
    db     *database.Database
}
```

### Key Functions

#### NewExchange

```go
func NewExchange(config *config.Config, db *database.Database) (Exchange, error)
```

Creates a new exchange instance. Currently returns a Binance exchange implementation.

#### GetPrice

```go
GetPrice(symbol string) (float64, error)
```

Fetches real-time price for a given trading pair.

#### PlaceOrder

```go
PlaceOrder(order *models.Order) error
```

Executes trades on the exchange with proper error handling and response parsing.

## Error Handling

- Network errors during API calls
- Invalid symbol errors
- Insufficient balance errors
- Order placement failures
- Database operation errors

## Usage Example

```go
exchange, err := exchange.NewExchange(cfg, db)
if err != nil {
    log.Fatal(err)
}

price, err := exchange.GetPrice("BTCUSDT")
if err != nil {
    log.Error("Failed to get price:", err)
}
```
