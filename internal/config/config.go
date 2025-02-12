package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	/* Load .env file if it exists */
	godotenv.Load()
}

type Config struct {
	BINANCE_API_KEY    string
	BINANCE_API_SECRET string
	InitialInvestment  float64
	MaxDrawdown        float64
	RiskPerTrade       float64
	AggressiveFactor   float64
	EnableCompounding  bool
	TradingPairs       []string
	DatabasePath       string
	TelegramToken      string
	TelegramChatID     string
	MinOrderSize       float64
}

/* Config from .env file */
func LoadConfig() (*Config, error) {
	cfg := &Config{
		BINANCE_API_KEY:    getEnvVar("BINANCE_API_KEY", ""),
		BINANCE_API_SECRET: getEnvVar("BINANCE_API_SECRET", ""),
		InitialInvestment:  getEnvFloatVar("INITIAL_INVESTMENT", 0), // default value of 0
		MaxDrawdown:        getEnvFloatVar("MAX_DRAWDOWN", 0),       // default value of 0
		RiskPerTrade:       getEnvFloatVar("RISK_PER_TRADE", 0),     // default value of 0
		TradingPairs:       []string{"TRADING_PAIRS", "BTCUSDT"},    // default value of BTCUSDT
		DatabasePath:       getEnvVar("DB_PATH", "data/trading_bot.db"),
		TelegramToken:      getEnvVar("TELEGRAM_TOKEN", ""),
		TelegramChatID:     getEnvVar("TELEGRAM_CHAT_ID", ""),
	}

	/* Validate required fields */
	if cfg.BINANCE_API_KEY == "" || cfg.BINANCE_API_SECRET == "" {
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
