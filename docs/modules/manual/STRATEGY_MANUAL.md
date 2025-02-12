# Trading Strategy Manual

## Table of Contents

1. [Overview](#overview)
2. [Mean Reversion Strategy](#mean-reversion-strategy)
3. [RSI Indicator](#rsi-indicator)
4. [Risk Management](#risk-management)
5. [Position Sizing](#position-sizing)
6. [Profit Mechanics](#profit-mechanics)

## Overview

This bot implements a mean reversion strategy using the Relative Strength Index (RSI) indicator. The strategy is based on the principle that prices tend to bounce back after reaching extreme levels.

## Mean Reversion Strategy

### Basic Concept

- When an asset is oversold (price too low), it's likely to rise
- When an asset is overbought (price too high), it's likely to fall
- We use RSI to identify these extreme conditions

### Trading Rules

1. **Buy Signal Conditions:**

   - RSI drops below 30 (oversold)
   - No existing position is open
   - Price is above minimum volume threshold

2. **Sell Signal Conditions:**
   - RSI rises above 70 (overbought)
   - Have an existing long position
   - Minimum profit target reached

### Example Scenario

```
Initial Price: $50,000
RSI drops to 28 → BUY Signal
Price rises to $52,000
RSI reaches 72 → SELL Signal
Profit: $2,000 (4% gain)
```

## RSI Indicator

### How RSI Works

```go
RSI = 100 - (100 / (1 + RS))
RS = Average Gain / Average Loss
```

### Parameters

- Period: 14 (standard setting)
- Oversold level: 30
- Overbought level: 70

### Interpretation

```
RSI > 70: Overbought → Potential Sell Signal
RSI < 30: Oversold → Potential Buy Signal
RSI = 50: Neutral → No Action
```

## Risk Management

### Position Sizing

```
Risk per trade = Account Balance × Risk Percentage
Position Size = Risk Amount / Stop Loss Distance
```

Example:

```
Account: $10,000
Risk per trade: 2% ($200)
Stop loss: 1% from entry
Position size = $200 / (Entry Price × 0.01)
```

### Stop Loss Placement

1. **Technical Stop Loss:**

   - 1% below entry price for long positions
   - Adjusted based on market volatility

2. **Maximum Drawdown Protection:**
   - Bot stops trading if drawdown exceeds 10%
   - Prevents catastrophic losses

## Position Sizing

### Formula

```
Position Size = (Account Balance × Risk Per Trade) / (Entry Price - Stop Loss)
```

### Example Calculation

```
Account Balance: $10,000
Risk Per Trade: 2% ($200)
Entry Price: $50,000
Stop Loss: $49,500 (1% below entry)
Position Size = $200 / $500 = 0.4 BTC
```

## Profit Mechanics

### Entry Strategy

1. Wait for RSI to drop below 30
2. Calculate position size based on risk
3. Place market buy order
4. Set stop loss order

### Exit Strategy

1. Primary Exit: RSI above 70
2. Take Profit Targets:
   - Partial exit at 2% profit
   - Full exit at 4% profit
3. Stop Loss: -1% from entry

### Profit Calculation Example

```
Entry: $50,000
Position Size: 0.4 BTC
Investment: $20,000

Scenario 1 (Win):
Exit at $52,000 (4% gain)
Profit = 0.4 × ($52,000 - $50,000) = $800

Scenario 2 (Loss):
Stop Loss at $49,500 (-1%)
Loss = 0.4 × ($49,500 - $50,000) = -$200
```

### Performance Metrics

- Win Rate Target: >60%
- Risk-Reward Ratio: 1:4
- Maximum Drawdown: 10%
- Expected Monthly Return: 5-15%

## Risk Warning

The strategy's performance depends on:

1. Market conditions (works best in ranging markets)
2. Proper risk management
3. Accurate execution of signals
4. Market liquidity
5. Trading fees and slippage

Remember:

- Past performance doesn't guarantee future results
- Always start with small amounts
- Monitor and adjust parameters based on performance
- Keep detailed trading logs for optimization

### Enhanced Entry Criteria

- **Strict Price Positioning**:
  - Buy only when price is in bottom 20% of recent range
  - Sell only when price is in top 20% of recent range
- **Multi-Indicator Confirmation**:
  - RSI < 30 + Price < Bollinger Lower Band → Strong Buy
  - RSI > 70 + Price > Bollinger Upper Band → Strong Sell
- **Trend Filter**:
  - Only trade when 50-period EMA is flat (±2% change)
  - Avoid trading in strong trends

### Smart Exit System

1. **Profit Protection**:
   - 50% position sold at 2% profit
   - 30% sold at 3% profit
   - 20% sold at 5% profit
2. **Trailing Stop**:
   - Activates after 1% profit
   - Tracks highest price reached
   - Triggers at 0.5% below peak
3. **Emergency Stop**:
   - Hard stop at 1% loss
   - Cancels all open orders
   - Enters cooldown period

## Profit Calculation Methodology

1. Each BUY creates a new position with unique PositionID
2. SELL orders must reference existing BUY positions
3. P&L calculated per completed position (BUY+SELL pair)
4. Realized P&L = (Sell Price - Buy Price) × Quantity
5. Unrealized P&L calculated for open positions

Example:

- BUY 1 BTC @ $50,000 → Position OPEN
- SELL 1 BTC @ $52,000 → Position CLOSED
- Realized P&L = ($52,000 - $50,000) × 1 = $2,000
