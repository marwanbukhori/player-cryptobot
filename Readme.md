# Crypto Trading Bot

A cryptocurrency trading bot built in Go using the Binance API. This bot implements various trading strategies with a focus on mean reversion trading.

## Features

- Real-time trading on Binance
- Mean reversion strategy with RSI indicator
- SQLite database for trade history
- Risk management system
- Configurable trading parameters
- Telegram notifications (optional)

## Project Structure

1. Structure of the project:

```bash
player-cryptobot/
├── cmd/ # Command line applications
│ └── bot/ # Main bot application
│ └── main.go # Entry point
├── internal/ # Private application code
│ ├── models/ # Data models
│ ├── config/ # Configuration management
│ ├── database/ # Database operations
│ ├── exchange/ # Exchange integration
│ └── strategy/ # Trading strategies
└── go.mod # Go module definition
```

2. Install dependencies:

```bash
go mod tidy
go get github.com/mattn/go-sqlite3
```

## Prerequisites

- Go 1.21 or higher
- Binance account with API access
- SQLite3

## Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/player-cryptobot.git
cd player-cryptobot
```

3. Set up environment variables:

```bash
export BINANCE_API_KEY="your_api_key"
export BINANCE_API_SECRET="your_api_secret"
export INITIAL_INVESTMENT="1000"
export MAX_DRAWDOWN="0.1"
export RISK_PER_TRADE="0.02"
```

## Configuration

The bot can be configured through environment variables:

- `BINANCE_API_KEY`: Your Binance API key
- `BINANCE_API_SECRET`: Your Binance API secret
- `INITIAL_INVESTMENT`: Initial investment amount
- `MAX_DRAWDOWN`: Maximum allowed drawdown (e.g., 0.1 for 10%)
- `RISK_PER_TRADE`: Risk per trade (e.g., 0.02 for 2%)
- `DB_PATH`: SQLite database path (default: "trading_bot.db")
- `TELEGRAM_TOKEN`: Telegram bot token (optional)
- `TELEGRAM_CHAT_ID`: Telegram chat ID (optional)

## Trading Strategies

### Mean Reversion Strategy

The bot implements a mean reversion strategy using the RSI (Relative Strength Index) indicator:

```go
type RSICalculator struct {
    period    int
    prevPrice float64
    gains     []float64
    losses    []float64
}
```

- Calculates RSI over a 14-period window
- Generates buy signals when RSI < 30 (oversold)
- Generates sell signals when RSI > 70 (overbought)

## Running the Bot

1. Build the bot:

```bash
go build -o trading-bot ./cmd/bot
```

2. Run the bot:

```bash
./trading-bot
```

## Docker Deployment

1. Build the Docker image:

```bash
docker build -t trading-bot .
```

2. Run the container:

```bash
docker run -d \
  --env BINANCE_API_KEY=xxx \
  --env BINANCE_API_SECRET=yyy \
  --env INITIAL_INVESTMENT=1000 \
  trading-bot
```

## Risk Management

The bot implements several risk management features:

- Maximum drawdown protection
- Position sizing based on risk per trade
- Stop-loss orders
- Trade history tracking

## Database Schema

The bot uses SQLite to store trade history:

```sql
CREATE TABLE trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL,
    price REAL NOT NULL,
    quantity REAL NOT NULL,
    timestamp DATETIME NOT NULL
)
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Disclaimer

This bot is for educational purposes only. Cryptocurrency trading carries significant risks. Use this bot at your own risk.
