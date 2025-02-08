package strategy

import (
	"testing"
)

func TestRSICalculator(t *testing.T) {
	rsi := NewRSICalculator(14)

	// Test cases
	prices := []float64{44.34, 44.09, 44.15, 43.61, 44.33}
	expected := []float64{0, 0, 0, 0, 51.78}

	for i, price := range prices {
		result := rsi.Calculate(price)
		if i >= len(expected) {
			continue
		}
		if result != expected[i] {
			t.Errorf("Expected RSI of %v, got %v", expected[i], result)
		}
	}
}
