package risk

import (
	"math"
)

type RiskManager struct {
	maxDrawdown       float64
	riskPerTrade      float64
	initialBalance    float64
	currentBalance    float64
	aggressiveFactor  float64
	enableCompounding bool
}

func NewRiskManager(initialBalance, maxDrawdown, riskPerTrade, aggressiveFactor float64, enableCompounding bool) *RiskManager {
	return &RiskManager{
		initialBalance:    initialBalance,
		currentBalance:    initialBalance,
		maxDrawdown:       maxDrawdown,
		riskPerTrade:      riskPerTrade,
		aggressiveFactor:  aggressiveFactor,
		enableCompounding: enableCompounding,
	}
}

/* Risk Manager is a component that manages the risk of the trading bot
*  - calculates the position size based on the current balance and the risk per trade
*  - checks if the balance is below the max drawdown
*  - updates the balance when a new trade is made
*  - gets the current balance
 */

/*
*  Calculate the position size
 */
func (r *RiskManager) CalculatePositionSize(price float64, stopLoss float64) (float64, error) {
	/*
	*  Use current balance instead of initial balance
	*  Cap at 2x initial
	 */
	riskBase := math.Min(r.currentBalance, r.initialBalance*2)
	riskAmount := riskBase * r.riskPerTrade

	/* Aggressive sizing during winning streaks */
	if r.currentBalance > r.initialBalance {
		riskAmount *= 1.5 // Risk 3% instead of 2% when profitable
	}

	stopLossDistance := math.Abs(price - stopLoss)
	quantity := riskAmount / stopLossDistance

	/* Dynamic max quantity based on current balance */
	maxQuantity := r.currentBalance / price
	return math.Min(quantity, maxQuantity), nil
}

/*
*  Update the balance
 */
func (r *RiskManager) UpdateBalance(newBalance float64) {
	r.currentBalance = newBalance
}

/*
*  Check if the balance is below the max drawdown
 */
func (r *RiskManager) CheckDrawdown() bool {
	if r.currentBalance <= 0 {
		return true
	}

	drawdown := (r.initialBalance - r.currentBalance) / r.initialBalance
	return drawdown >= r.maxDrawdown
}

/*
*  Get the current balance
 */
func (r *RiskManager) GetCurrentBalance() float64 {
	return r.currentBalance
}
