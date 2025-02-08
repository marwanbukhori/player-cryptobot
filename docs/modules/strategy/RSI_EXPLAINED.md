# RSI (Relative Strength Index) Explained

## Overview

RSI is a momentum oscillator that measures the speed and magnitude of price changes. It oscillates between 0 and 100.

## Key Levels

```
Overbought: RSI > 70 (Price might be too high, potential sell)
Neutral:    RSI 30-70 (Normal trading range)
Oversold:   RSI < 30 (Price might be too low, potential buy)
```

## Calculation Steps

1. **Price Changes**

```go
change = currentPrice - previousPrice

if change > 0:
    gain = change
    loss = 0
else:
    gain = 0
    loss = |change|
```

2. **Average Gains and Losses**

```go
// Over 14 periods (standard)
averageGain = sum(gains) / 14
averageLoss = sum(losses) / 14
```

3. **Relative Strength (RS)**

```go
RS = averageGain / averageLoss
```

4. **RSI Formula**

```go
RSI = 100 - (100 / (1 + RS))
```

## Example Calculation

```
Price Data:
Day 1: $100 → $105 (gain: 5)
Day 2: $105 → $103 (loss: 2)
Day 3: $103 → $107 (gain: 4)

averageGain = (5 + 0 + 4) / 3 = 3
averageLoss = (0 + 2 + 0) / 3 = 0.67

RS = 3 / 0.67 = 4.48
RSI = 100 - (100 / (1 + 4.48)) = 81.75
```

## Trading Psychology

1. **Overbought (RSI > 70)**

```
- Market is potentially overextended
- Buyers might be exhausted
- Price might reverse downward
- Generate SELL signal
```

2. **Oversold (RSI < 30)**

```
- Market is potentially undervalued
- Sellers might be exhausted
- Price might reverse upward
- Generate BUY signal
```

## Implementation Details

1. **Data Collection**

```go
type RSICalculator struct {
    period    int      // Usually 14 days
    gains     []float64
    losses    []float64
}
```

2. **Sliding Window**

```go
// Keep only last 14 values
if len(gains) > period {
    gains = gains[len(gains)-period:]
    losses = losses[len(losses)-period:]
}
```

3. **Signal Generation**

```go
func Analyze(price float64) *Signal {
    rsi := Calculate(price)

    if rsi < 30 {
        return BUY
    }
    if rsi > 70 {
        return SELL
    }

    return nil
}
```

## Advantages & Limitations

### Advantages

```
+ Clear buy/sell signals
+ Measures momentum
+ Works in ranging markets
+ Easy to understand
```

### Limitations

```
- Can stay overbought/oversold in strong trends
- Needs enough historical data (14 periods)
- May give false signals in trending markets
```

## Best Practices

1. **Confirmation**

```
- Wait for RSI to cross back above 30 for buys
- Wait for RSI to cross back below 70 for sells
- Consider trend direction
```

2. **Risk Management**

```
- Don't rely solely on RSI
- Use stop losses
- Consider position sizing
- Monitor market conditions
```
