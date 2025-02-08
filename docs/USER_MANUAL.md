# Crypto Trading Bot - User Manual

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Running the Bot](#running-the-bot)
5. [Backtesting](#backtesting)
6. [Monitoring](#monitoring)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

- Go 1.21 or higher
- Binance account with API access
- Telegram account (optional, for notifications)
- SQLite3

## Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/player-cryptobot.git
cd player-cryptobot
```

2. Install dependencies:

```bash
go mod tidy
go get github.com/mattn/go-sqlite3
go get github.com/joho/godotenv
```

## Configuration

1. Create a `.env` file in the root directory:

```bash
cp .env.example .env
```

2. Configure your Binance API (from https://www.binance.com/en/my/settings/api-management):

```env
BINANCE_API_KEY="your_api_key"
BINANCE_API_SECRET="your_api_secret"
```

3. Set up Telegram notifications (optional):
   - Create a Telegram bot via @BotFather
   - Get your chat ID by:
     1. Send a message to your bot
     2. Visit: https://api.telegram.org/bot<YourBOTToken>/getUpdates
     3. Copy the "chat":{"id":XXXXXXXXX} number
   - Add to .env:

```env
TELEGRAM_TOKEN="your_bot_token"
TELEGRAM_CHAT_ID="your_chat_id"
```

4. Configure trading parameters:

```env
INITIAL_INVESTMENT=1000    # Starting amount in USDT
MAX_DRAWDOWN=0.1          # Maximum allowed loss (10%)
RISK_PER_TRADE=0.02       # Risk per trade (2%)
```

## Running the Bot

1. Start the bot:

```bash
go run cmd/bot/main.go
```

2. Monitor the logs:

- INFO: General information about bot operation
- ERROR: Issues that need attention
- DEBUG: Detailed information about prices and analysis

3. The bot will:
   - Monitor configured trading pairs
   - Calculate RSI values
   - Generate buy signals when RSI < 30 (oversold)
   - Generate sell signals when RSI > 70 (overbought)
   - Manage position sizes based on risk parameters
   - Send notifications via Telegram (if configured)

## Backtesting

1. Prepare historical data in CSV format:

```csv
timestamp,price,volume
2024-01-01 00:00:00,50000,100
2024-01-01 00:01:00,50100,95
...
```

2. Run backtest:

```bash
go run cmd/backtest/main.go -data path/to/your/data.csv -balance 10000
```

3. Analyze results:

- Total trades executed
- Win rate
- Total profit/loss
- Maximum drawdown

## Monitoring

1. Telegram Notifications:

- Trade executions
- Error alerts
- System status updates

2. Database Records:

- All trades are stored in SQLite database
- View trade history:

```bash
sqlite3 trading_bot.db
SELECT * FROM trades ORDER BY timestamp DESC LIMIT 10;
```

## Troubleshooting

Common Issues:

1. "Binance API key and secret are required"

   - Check your .env file
   - Verify API key permissions on Binance

2. "Error getting price"

   - Check internet connection
   - Verify trading pair exists

3. "Insufficient balance"

   - Check INITIAL_INVESTMENT setting
   - Verify Binance account balance

4. No Telegram notifications
   - Verify TELEGRAM_TOKEN and CHAT_ID
   - Ensure bot is started in Telegram

## Safety Recommendations

1. Start with small amounts
2. Monitor the bot regularly
3. Keep API keys secure
4. Use IP restrictions on Binance API
5. Regular backup of trading_bot.db

## Disclaimer

This bot is for educational purposes. Cryptocurrency trading carries significant risks. Use at your own risk.

## Trading Frequency & Order Types

- **Typical Daily Trades**: 3-5
- **Order Execution**:
  - Market orders for entries
  - Limit orders for profit taking
- **Cooldown Periods**:
  - 30 minutes after loss
  - 15 minutes after win
- **Lottery Protection**:
  - Max 2 trades/hour
  - Max 10 trades/day
  - Weekend trading disabled
