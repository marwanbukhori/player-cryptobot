# Risk Management Module Documentation

## Overview

The risk management module handles position sizing, drawdown protection, and overall risk control.

## Components

### RiskManager

```go
type RiskManager struct {
    maxDrawdown    float64
    riskPerTrade   float64
    initialBalance float64
    currentBalance float64
}
```

### Key Functions

#### CalculatePositionSize

```go
func (r *RiskManager) CalculatePositionSize(price float64, stopLoss float64) (float64, error)
```

Calculates safe position size based on:

- Account balance
- Risk per trade
- Stop loss distance

#### CheckDrawdown

```go
func (r *RiskManager) CheckDrawdown() bool
```

Monitors drawdown against maximum allowed drawdown.

### Risk Parameters

- Maximum Drawdown: Configurable (default 10%)
- Risk Per Trade: Configurable (default 2%)
- Stop Loss: Dynamic based on strategy

## Usage Example

```go
riskManager := risk.NewRiskManager(10000.0, 0.1, 0.02)
size, err := riskManager.CalculatePositionSize(price, stopLoss)
```
