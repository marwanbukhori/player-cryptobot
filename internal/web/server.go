package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/marwanbukhori/player-cryptobot/internal/exchange"
)

type Server struct {
	exchange exchange.Exchange
	port     string
	tmpl     *template.Template
}

type DashboardData struct {
	Summary []TradingSummary `json:"summary"`
	Trades  []TradeData      `json:"trades"`
}

type TradingSummary struct {
	Symbol        string    `json:"symbol"`
	TotalTrades   int       `json:"total_trades"`
	WinningTrades int       `json:"winning_trades"`
	LosingTrades  int       `json:"losing_trades"`
	TotalPnL      float64   `json:"total_pnl"`
	AvgPnLPercent float64   `json:"avg_pn_l_percent"`
	TotalVolume   float64   `json:"total_volume"`
	FirstTrade    time.Time `json:"first_trade"`
	LastTrade     time.Time `json:"last_trade"`
}

type TradeData struct {
	ID         uint      `json:"id"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`
	Price      float64   `json:"price"`
	Quantity   float64   `json:"quantity"`
	Value      float64   `json:"value"`
	Fee        float64   `json:"fee"`
	Timestamp  time.Time `json:"timestamp"`
	PnL        float64   `json:"pn_l"`
	PnLPercent float64   `json:"pn_l_percent"`
	Status     string    `json:"status"`
}

func NewServer(exchange exchange.Exchange, port string) *Server {
	funcMap := template.FuncMap{
		"lower": strings.ToLower,
		"div": func(a, b int, scale float64) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b) * scale
		},
	}

	tmpl := template.Must(template.New("dashboard").Funcs(funcMap).Parse(dashboardTemplate))
	return &Server{
		exchange: exchange,
		port:     port,
		tmpl:     tmpl,
	}
}

func (s *Server) Start() error {
	// Serve static files
	fs := http.FileServer(http.Dir("internal/web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Add favicon handler
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Routes
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/trades", s.handleTrades)
	http.HandleFunc("/api/summary", s.handleSummary)
	http.HandleFunc("/export/csv", s.handleExportCSV)
	http.HandleFunc("/export/json", s.handleExportJSON)

	fmt.Printf("Starting web server on %s\n", s.port)
	return http.ListenAndServe(s.port, nil)
}

const dashboardTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Trading Bot Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        .card { margin-bottom: 20px; }
        .trade-row.buy { background-color: rgba(40, 167, 69, 0.1); }
        .trade-row.sell { background-color: rgba(220, 53, 69, 0.1); }
    </style>
</head>
<body>
    <div class="container mt-4">
        <h1 class="mb-4">Trading Bot Dashboard</h1>

        <!-- Trading Summary -->
        <div class="row">
            {{range .Summary}}
            <div class="col-md-4">
                <div class="card">
                    <div class="card-body">
                        <h5 class="card-title">{{.Symbol}}</h5>
                        <p class="card-text">
                            Total Trades: {{.TotalTrades}}<br>
                            Win Rate: {{if .TotalTrades}}{{printf "%.2f" (div .WinningTrades .TotalTrades 100)}}%{{else}}0%{{end}}<br>
                            Total P&L: {{printf "%.2f" .TotalPnL}} USDT<br>
                            Avg P&L: {{printf "%.2f" .AvgPnLPercent}}%<br>
                            Volume: {{printf "%.2f" .TotalVolume}} USDT
                        </p>
                    </div>
                </div>
            </div>
            {{end}}
        </div>

        <!-- Recent Trades -->
        <div class="card">
            <div class="card-body">
                <h5 class="card-title">Recent Trades</h5>
                <div class="table-responsive">
                    <table class="table table-striped">
                        <thead>
                            <tr>
                                <th>Time</th>
                                <th>Symbol</th>
                                <th>Side</th>
                                <th>Price</th>
                                <th>Quantity</th>
                                <th>Value</th>
                                <th>P&L</th>
                                <th>Status</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .Trades}}
                            <tr class="trade-row {{.Side | lower}}">
                                <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
                                <td>{{.Symbol}}</td>
                                <td>{{.Side}}</td>
                                <td>{{printf "%.2f" .Price}}</td>
                                <td>{{printf "%.8f" .Quantity}}</td>
                                <td>{{printf "%.2f" .Value}}</td>
                                <td>{{if eq .Side "SELL"}}{{printf "%.2f" .PnL}}{{end}}</td>
                                <td>{{.Status}}</td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        // Auto-refresh every 30 seconds
        setInterval(() => {
            location.reload();
        }, 30000);
    </script>
</body>
</html>
`
