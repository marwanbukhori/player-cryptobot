# Strategy Module Documentation

## Overview

The strategy module implements various trading strategies, with the primary focus on mean reversion using RSI.

## Components

### Strategy Interface

```go
type Strategy interface {
    Analyze(data *models.MarketData) *models.Signal
}
```

### Mean Reversion Strategy

Implements mean reversion using RSI (Relative Strength Index).

#### RSICalculator

```go
type RSICalculator struct {
    period    int
    prevPrice float64
    gains     []float64
    losses    []float64
}
```

### Parameters

- RSI Period: 14 (default)
- Oversold Level: 30
- Overbought Level: 70

### Trading Logic

1. Calculate RSI value
2. Generate buy signals when RSI < 30 (oversold)
3. Generate sell signals when RSI > 70 (overbought)

## Testing

The strategy includes unit tests with known price sequences and expected RSI values.

## Usage Example

```go
strategy := strategy.NewMeanReversionStrategy()
signal := strategy.Analyze(&models.MarketData{
    Symbol: "BTCUSDT",
    Price:  price,
    Time:   time.Now(),
})
```
