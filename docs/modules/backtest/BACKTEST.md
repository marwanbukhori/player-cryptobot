# Backtesting Module Documentation

## Overview

The backtesting module allows testing trading strategies against historical data.

## Components

### Backtester

```go
type Backtester struct {
    strategy strategy.Strategy
    data     []MarketData
    balance  float64
}
```

### BacktestResult

```go
type BacktestResult struct {
    TotalTrades   int
    WinningTrades int
    LosingTrades  int
    TotalProfit   float64
    MaxDrawdown   float64
    WinRate       float64
    Trades        []models.Order
}
```

### Key Functions

#### Run

```go
func (b *Backtester) Run() BacktestResult
```

Executes backtest and returns performance metrics.

#### LoadData

```go
func (b *Backtester) LoadData(data []MarketData)
```

Loads historical market data for testing.

## Metrics Calculated

- Total number of trades
- Win rate
- Total profit/loss
- Maximum drawdown
- Trade history

## Usage Example

```go
tester := backtest.NewBacktester(strategy, 10000.0)
tester.LoadData(historicalData)
result := tester.Run()
```
