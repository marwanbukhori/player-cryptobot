# Mean Reversion Strategy with RSI

## Core Concept

The strategy is based on the idea that prices tend to "revert to the mean" - when they go too high or too low, they'll likely bounce back.

## RSI (Relative Strength Index)

- Period: 14 (standard setting)
- Range: 0 to 100
- Formula: RSI = 100 - (100 / (1 + RS))
  where RS = Average Gains / Average Losses

## Trading Signals

### Buy Signal (Oversold)

```
When RSI < 30:
- Market is considered oversold
- Price likely to go up
- Bot generates BUY signal
```

### Sell Signal (Overbought)

```
When RSI > 70:
- Market is considered overbought
- Price likely to go down
- Bot generates SELL signal
```

## Risk Management

- Initial Investment: 10 USDT
- Max Drawdown: 10%
- Risk per Trade: 2%

## Example Scenario

```
BTC Price: $50,000
RSI drops to 28 → OVERSOLD
Bot Action: BUY
...
Price rises to $52,000
RSI reaches 72 → OVERBOUGHT
Bot Action: SELL
Result: ~4% profit
```
