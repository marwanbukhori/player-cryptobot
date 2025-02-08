package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/marwanbukhori/player-cryptobot/internal/backtest"
	"github.com/marwanbukhori/player-cryptobot/internal/config"
	"github.com/marwanbukhori/player-cryptobot/internal/database"
	"github.com/marwanbukhori/player-cryptobot/internal/exchange"
	"github.com/marwanbukhori/player-cryptobot/internal/strategy"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize exchange
	exchange, err := exchange.NewExchange(cfg, db)
	if err != nil {
		log.Fatal("Failed to initialize exchange:", err)
	}
	strategy := strategy.NewMeanReversionStrategy()

	// Get historical data
	data, err := exchange.GetHistoricalData("BTCUSDT", "1m", 1000) // Last 1000 minutes
	if err != nil {
		log.Fatal(err)
	}

	// Run backtest
	results := backtest.Run(data, strategy, 10.0) // Start with 10 USDT

	// Print results
	fmt.Printf("Total Trades: %d\n", results.TotalTrades)
	fmt.Printf("Win Rate: %.2f%%\n", results.WinRate)
	fmt.Printf("Profit/Loss: %.2f USDT\n", results.ProfitLoss)
}

func loadHistoricalData(filepath string) ([]backtest.MarketData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var data []backtest.MarketData
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		timestamp, _ := time.Parse("2006-01-02 15:04:05", record[0])
		price, _ := strconv.ParseFloat(record[1], 64)
		volume, _ := strconv.ParseFloat(record[2], 64)

		data = append(data, backtest.MarketData{
			Time:   timestamp,
			Price:  price,
			Volume: volume,
		})
	}

	return data, nil
}
