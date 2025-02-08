package web

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/marwanbukhori/player-cryptobot/internal/models"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	summary, err := s.exchange.GetTradingSummary()
	if err != nil {
		http.Error(w, "Failed to get trading summary", http.StatusInternalServerError)
		return
	}

	trades, err := s.exchange.GetRecentTrades(10)
	if err != nil {
		http.Error(w, "Error fetching trades", http.StatusInternalServerError)
		return
	}

	// Calculate cumulative P&L for chart
	var labels []string
	var pnlData []float64
	cumulativePnL := 0.0

	for i := len(trades) - 1; i >= 0; i-- {
		t := trades[i]
		if t.Side == "SELL" {
			cumulativePnL += t.PnL
			labels = append(labels, t.Timestamp.Format("2006-01-02 15:04"))
			pnlData = append(pnlData, cumulativePnL)
		}
	}

	data := struct {
		TotalTrades  int
		TotalVolume  float64
		TotalPnL     float64
		WinRate      float64
		RecentTrades []models.Trade
		Labels       []string
		PnLData      []float64
	}{
		RecentTrades: trades,
		Labels:       labels,
		PnLData:      pnlData,
	}

	if len(summary) > 0 {
		s := summary[0]
		data.TotalTrades = s.TotalTrades
		data.TotalPnL = s.TotalPnL
		if s.TotalTrades > 0 {
			data.WinRate = float64(s.WinningTrades) / float64(s.TotalTrades) * 100
		}
		data.TotalVolume = s.TotalVolume
	}

	tmpl := template.Must(template.ParseFiles("internal/web/templates/dashboard.html"))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	if err := csvWriter.Write([]string{"Date", "Symbol", "Side", "Price", "Quantity", "Value", "Fee", "PnL"}); err != nil {
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
		}); err != nil {
			http.Error(w, "Failed to write CSV data", http.StatusInternalServerError)
			return
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		http.Error(w, "Failed to flush CSV data", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleTrades(w http.ResponseWriter, r *http.Request) {
	trades, err := s.exchange.GetAllTrades()
	if err != nil {
		http.Error(w, "Failed to get trades", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
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
