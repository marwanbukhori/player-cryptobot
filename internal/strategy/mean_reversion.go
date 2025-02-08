package strategy

import (
	"github.com/marwanbukhori/player-cryptobot/internal/models"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type MeanReversionStrategy struct {
	rsi        *RSICalculator
	lastPrices map[string][]float64 // Track price history per symbol
	maxPrices  map[string]float64   // Track local highs
	minPrices  map[string]float64   // Track local lows
}

func NewMeanReversionStrategy() *MeanReversionStrategy {
	return &MeanReversionStrategy{
		rsi:        NewRSICalculator(5),
		lastPrices: make(map[string][]float64),
		maxPrices:  make(map[string]float64),
		minPrices:  make(map[string]float64),
	}
}

func (s *MeanReversionStrategy) Analyze(data *models.MarketData) *models.Signal {
	// Track prices
	prices := s.lastPrices[data.Symbol]
	prices = append(prices, data.Price)
	if len(prices) > 30 { // Keep last 30 minutes
		prices = prices[1:]
	}
	s.lastPrices[data.Symbol] = prices

	// Update local highs and lows
	if s.maxPrices[data.Symbol] < data.Price {
		s.maxPrices[data.Symbol] = data.Price
	}
	if s.minPrices[data.Symbol] == 0 || s.minPrices[data.Symbol] > data.Price {
		s.minPrices[data.Symbol] = data.Price
	}

	// Calculate price metrics
	localHigh := s.maxPrices[data.Symbol]
	localLow := s.minPrices[data.Symbol]
	priceRange := localHigh - localLow

	// Calculate position in range (0-100%)
	positionInRange := 0.0
	if priceRange > 0 {
		positionInRange = ((data.Price - localLow) / priceRange) * 100
	}

	rsi := s.rsi.Calculate(data.Price)

	log.Infof("Symbol: %s, Price: %.2f, RSI: %.2f, Range Position: %.2f%%",
		data.Symbol, data.Price, rsi, positionInRange)

	// Trading logic
	if positionInRange < 20 && rsi < 40 { // Price near bottom + oversold
		log.Infof("BUY SIGNAL - %s: Price near low (%.2f%%) and RSI oversold (%.2f)",
			data.Symbol, positionInRange, rsi)
		return &models.Signal{Symbol: data.Symbol, Action: "BUY"}
	}

	if positionInRange > 80 && rsi > 60 { // Price near top + overbought
		log.Infof("SELL SIGNAL - %s: Price near high (%.2f%%) and RSI overbought (%.2f)",
			data.Symbol, positionInRange, rsi)
		return &models.Signal{Symbol: data.Symbol, Action: "SELL"}
	}

	return nil
}

// RSICalculator calculates the Relative Strength Index
type RSICalculator struct {
	period    int
	prevPrice float64
	gains     []float64
	losses    []float64
}

func NewRSICalculator(period int) *RSICalculator {
	return &RSICalculator{
		period: period,
		gains:  make([]float64, 0, period),
		losses: make([]float64, 0, period),
	}
}

func (r *RSICalculator) Calculate(price float64) float64 {
	if r.prevPrice == 0 {
		r.prevPrice = price
		return 50
	}

	// More sensitive percentage change calculation
	change := ((price - r.prevPrice) / r.prevPrice) * 1000 // Multiplied by 1000 instead of 100
	r.prevPrice = price

	// Store changes
	if change > 0 {
		r.gains = append(r.gains, change)
		r.losses = append(r.losses, 0)
	} else {
		r.gains = append(r.gains, 0)
		r.losses = append(r.losses, -change)
	}

	// Keep only the last 'period' values
	if len(r.gains) > r.period {
		r.gains = r.gains[1:] // Remove oldest value
		r.losses = r.losses[1:]
	}

	// Need enough data points
	if len(r.gains) < r.period {
		return 50
	}

	avgGain := average(r.gains)
	avgLoss := average(r.losses)

	if avgLoss == 0 {
		if avgGain == 0 {
			return 50
		}
		return 100
	}

	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
