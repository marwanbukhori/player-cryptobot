package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file if it exists
	godotenv.Load()
}

type Config struct {
	APIKey            string
	APISecret         string
	InitialInvestment float64
	MaxDrawdown       float64
	RiskPerTrade      float64
	AggressiveFactor  float64
	EnableCompounding bool
	TradingPairs      []string
	DatabasePath      string
	TelegramToken     string
	TelegramChatID    string
	MinOrderSize      float64
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		APIKey:            getEnvVar("BINANCE_API_KEY", ""),
		APISecret:         getEnvVar("BINANCE_API_SECRET", ""),
		InitialInvestment: getEnvFloatVar("INITIAL_INVESTMENT", 10.0),
		MaxDrawdown:       getEnvFloatVar("MAX_DRAWDOWN", 0.1),    // 10%
		RiskPerTrade:      getEnvFloatVar("RISK_PER_TRADE", 0.02), // 2%
		TradingPairs:      []string{"BTCUSDT"},                    // Updated pairs
		DatabasePath:      getEnvVar("DB_PATH", "data/trading_bot.db"),
		TelegramToken:     getEnvVar("TELEGRAM_TOKEN", ""),
		TelegramChatID:    getEnvVar("TELEGRAM_CHAT_ID", ""),
	}

	// Validate required fields
	if cfg.APIKey == "" || cfg.APISecret == "" {
		return nil, fmt.Errorf("Binance API key and secret are required")
	}

	minOrderSize, _ := strconv.ParseFloat(os.Getenv("MIN_ORDER_SIZE"), 64)
	if minOrderSize == 0 {
		minOrderSize = 10 // Default minimum order size in USDT
	}

	cfg.MinOrderSize = minOrderSize

	return cfg, nil
}

func getEnvVar(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvFloatVar(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
