package web

import (
	"net/http"

	"github.com/marwanbukhori/player-cryptobot/internal/exchange"
)

type Server struct {
	exchange exchange.Exchange
	port     string
}

func NewServer(exchange exchange.Exchange, port string) *Server {
	return &Server{
		exchange: exchange,
		port:     port,
	}
}

func (s *Server) Start() error {
	// Add favicon handler
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Routes
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/trades", s.handleTrades)
	http.HandleFunc("/export/csv", s.handleExportCSV)
	http.HandleFunc("/export/json", s.handleExportJSON)

	return http.ListenAndServe(s.port, nil)
}
