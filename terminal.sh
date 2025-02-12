# Ignore this. For learning purposes only.

# Create the directory structure
mkdir -p cmd/bot
mkdir -p internal/exchange
mkdir -p internal/strategy
mkdir -p internal/risk
mkdir -p internal/config
mkdir -p internal/models
mkdir -p internal/database
mkdir -p docs/modules
mkdir -p data

# Initialize the Go module (if not already done)
go mod init github.com/marwanbukhori/player-cryptobot

# Get dependencies
go mod tidy
go get github.com/mattn/go-sqlite3
go get github.com/joho/godotenv

touch .env

# For Ubuntu/Debian
sudo apt-get install sqlite3

# For macOS
brew install sqlite3

# For Windows
# Download from https://www.sqlite.org/download.html

sqlite3 data/trading_bot.db

# Connect to database
sqlite3 data/trading_bot.db

# Show tables
.tables

# Show schema
.schema trades

# Show recent trades
SELECT * FROM trades ORDER BY timestamp DESC LIMIT 5;

# Show trading statistics
SELECT
    COUNT(*) as total_trades,
    SUM(CASE WHEN price > 0 THEN 1 ELSE 0 END) as winning_trades,
    SUM(quantity * price) as total_volume
FROM trades;

# Exit
.quit

# Backup database
sqlite3 data/trading_bot.db ".backup 'data/trading_bot_backup.db'"

# Optimize database
sqlite3 data/trading_bot.db "VACUUM;"

# Export to CSV
sqlite3 -header -csv data/trading_bot.db "SELECT * FROM trades;" > trades_export.csv
