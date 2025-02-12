# Trading Bot Startup Guide

## Table of Contents

1. [Pre-Flight Checklist](#pre-flight-checklist)
2. [Configuration Setup](#configuration-setup)
3. [Starting the Bot](#starting-the-bot)
4. [Monitoring Trades](#monitoring-trades)
5. [Safety Guidelines](#safety-guidelines)

## Pre-Flight Checklist

### 1. Environment Verification

- [ ] All dependencies installed
- [ ] Database initialized
- [ ] `.env` file configured
- [ ] Telegram bot set up (optional)

### 2. API Configuration

```bash
# Test Binance API connection
curl -H "X-MBX-APIKEY: your_api_key" https://api.binance.com/api/v3/account
```

Verify API permissions:

- [ ] Reading enabled
- [ ] Spot trading enabled
- [ ] IP restrictions set (recommended)

### 3. Risk Parameters

Check `.env` configuration:

```env
INITIAL_INVESTMENT=10                # Start small!
MAX_DRAWDOWN=0.1                    # 10% maximum loss
RISK_PER_TRADE=0.02                 # 2% per trade
```

## Configuration Setup

### 1. Trading Pairs

Default pairs in `config.go`:

```go
TradingPairs: []string{"BTCUSDT", "ETHUSDT"}
```

### 2. Strategy Parameters

Mean Reversion Strategy settings:

- RSI Period: 14
- Oversold level: 30 (buy signal)
- Overbought level: 70 (sell signal)

## Starting the Bot

### 1. Initial Start (Paper Trading)

```bash
# Start the bot
go run cmd/bot/main.go
```

### 2. Monitor Initial Output

Look for:

```
INFO: Bot started successfully
INFO: Connected to Binance
INFO: Database initialized
INFO: Telegram notifications enabled
```

### 3. First Trade Verification

```bash
# Check database for trades
sqlite3 data/trading_bot.db "SELECT * FROM trades ORDER BY timestamp DESC LIMIT 1;"
```

## Monitoring Trades

### 1. Real-Time Monitoring

- Watch Telegram notifications
- Monitor log output
- Check database entries

### 2. Performance Tracking

```bash
# Get trading statistics
sqlite3 data/trading_bot.db "
SELECT
    COUNT(*) as total_trades,
    SUM(CASE WHEN price > 0 THEN 1 ELSE 0 END) as winning_trades,
    SUM(quantity * price) as total_volume
FROM trades;"
```

### 3. Error Checking

Common error messages and actions:

- "Insufficient balance" → Check account funds
- "API rate limit" → Reduce polling frequency
- "Order rejected" → Check price/quantity limits

## Safety Guidelines

### 1. Starting Small

1. Begin with minimum investment
2. Monitor few trading pairs initially
3. Verify all trades are as expected

### 2. Risk Management

- Never modify MAX_DRAWDOWN during active trades
- Keep emergency stop procedure ready
- Monitor account balance regularly

### 3. Emergency Procedures

To stop the bot:

```bash
# Graceful shutdown
Ctrl + C

# Check for open orders
curl -H "X-MBX-APIKEY: your_api_key" https://api.binance.com/api/v3/openOrders
```

## Daily Operations

### 1. Morning Checklist

- [ ] Check bot status
- [ ] Verify database connectivity
- [ ] Review overnight trades
- [ ] Check account balance

### 2. Monitoring Schedule

```bash
# Every 4 hours
- Check trading statistics
- Verify profit/loss
- Monitor error logs

# Daily
- Backup database
- Review performance metrics
- Check system resources
```

### 3. Performance Review

```sql
-- Daily P&L
SELECT
    DATE(timestamp) as trade_date,
    COUNT(*) as trades,
    SUM(quantity * price) as volume
FROM trades
GROUP BY DATE(timestamp)
ORDER BY trade_date DESC;
```

## Troubleshooting

### Common Issues

1. **No Trades Executing**

- Check RSI values
- Verify price feeds
- Confirm account balance

2. **Frequent Errors**

- Review API rate limits
- Check network connectivity
- Verify order size limits

3. **Performance Issues**

- Monitor system resources
- Check database size
- Review log file size

## Best Practices

1. **Risk Management**

- Never risk more than you can afford to lose
- Keep majority of funds in cold storage
- Regular profit taking

2. **System Health**

- Regular database maintenance
- Log rotation
- System time synchronization

3. **Documentation**

- Keep trade logs
- Document configuration changes
- Track system modifications

## Disclaimer

This bot is for educational purposes. Cryptocurrency trading carries significant risks. Always start with small amounts and never trade more than you can afford to lose.
