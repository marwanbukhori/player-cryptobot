package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

func main() {
	router := gin.Default()

	// Trading endpoints
	router.GET("/api/status", getBotStatus)
	router.POST("/api/trade", executeTrade)
	router.GET("/api/metrics", getPerformanceMetrics)

	// Frontend
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")
	router.GET("/", dashboardHandler)

	router.Run(":8080")
}

func dashboardHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"Title": "Trading Bot Dashboard",
	})
}

func getBotStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "running",
		"uptime": time.Since(startTime).String(),
	})
}

func executeTrade(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Trade executed",
	})
}

func getPerformanceMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_trades": 0,
		"win_rate":     0,
		"pnl":          0,
	})
}
