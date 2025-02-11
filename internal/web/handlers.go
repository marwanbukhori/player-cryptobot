package web

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/marwanbukhori/player-cryptobot/internal/models"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	summary, err := s.exchange.GetTradingSummary()
	if err != nil {
		// Don't return error, just log it and continue with empty data
		summary = []models.TradingSummary{}
	}

	trades, err := s.exchange.GetRecentTrades(100)
	if err != nil {
		// Don't return error, just log it and continue with empty data
		trades = []models.Trade{}
	}

	data := DashboardData{
		Summary: make([]TradingSummary, len(summary)),
		Trades:  make([]TradeData, len(trades)),
	}

	// Convert summary
	for i, s := range summary {
		data.Summary[i] = TradingSummary{
			Symbol:        s.Symbol,
			TotalTrades:   s.TotalTrades,
			WinningTrades: s.WinningTrades,
			LosingTrades:  s.LosingTrades,
			TotalPnL:      s.TotalPnL,
			AvgPnLPercent: s.AvgPnLPercent,
			TotalVolume:   s.TotalVolume,
			FirstTrade:    s.FirstTrade,
			LastTrade:     s.LastTrade,
		}
	}

	// Convert trades
	for i, t := range trades {
		data.Trades[i] = TradeData{
			ID:         t.ID,
			Symbol:     t.Symbol,
			Side:       t.Side,
			Price:      t.Price,
			Quantity:   t.Quantity,
			Value:      t.Value,
			Fee:        t.Fee,
			Timestamp:  t.Timestamp,
			PnL:        t.PnL,
			PnLPercent: t.PnLPercent,
			Status:     t.Status,
		}
	}

	if err := s.tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleTrades(w http.ResponseWriter, r *http.Request) {
	trades, err := s.exchange.GetRecentTrades(100)
	if err != nil {
		http.Error(w, "Failed to get trades", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.exchange.GetTradingSummary()
	if err != nil {
		http.Error(w, "Failed to get summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	trades, err := s.exchange.GetAllTrades()
	if err != nil {
		http.Error(w, "Failed to get trades", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=trades.csv")

	csvWriter := csv.NewWriter(w)
	if err := csvWriter.Write([]string{
		"Date", "Symbol", "Side", "Price", "Quantity", "Value", "Fee", "PnL", "PnL%", "Status",
	}); err != nil {
		http.Error(w, "Failed to write CSV header", http.StatusInternalServerError)
		return
	}

	for _, t := range trades {
		if err := csvWriter.Write([]string{
			t.Timestamp.Format(time.RFC3339),
			t.Symbol,
			t.Side,
			fmt.Sprintf("%.8f", t.Price),
			fmt.Sprintf("%.8f", t.Quantity),
			fmt.Sprintf("%.8f", t.Value),
			fmt.Sprintf("%.8f", t.Fee),
			fmt.Sprintf("%.8f", t.PnL),
			fmt.Sprintf("%.2f", t.PnLPercent),
			t.Status,
		}); err != nil {
			http.Error(w, "Failed to write CSV data", http.StatusInternalServerError)
			return
		}
	}
	csvWriter.Flush()
}

func (s *Server) handleExportJSON(w http.ResponseWriter, r *http.Request) {
	trades, err := s.exchange.GetAllTrades()
	if err != nil {
		http.Error(w, "Failed to get trades", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment;filename=trades.json")
	if err := json.NewEncoder(w).Encode(trades); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}
