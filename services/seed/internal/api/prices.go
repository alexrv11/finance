package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/prices?symbol=AAPL&interval=1d&from=2024-01-01&to=2024-12-31
func (s *Server) getPrices(c *gin.Context) {
	symbol := c.Query("symbol")
	interval := c.Query("interval")
	if symbol == "" || interval == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol and interval are required"})
		return
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bars, err := s.db.QueryBars(c.Request.Context(), symbol, interval, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type barDTO struct {
		TS       time.Time `json:"ts"`
		Open     float64   `json:"open"`
		High     float64   `json:"high"`
		Low      float64   `json:"low"`
		Close    float64   `json:"close"`
		Volume   int64     `json:"volume"`
		AdjClose float64   `json:"adj_close,omitempty"`
		Source   string    `json:"source"`
	}

	out := make([]barDTO, len(bars))
	for i, b := range bars {
		out[i] = barDTO{
			TS: b.TS, Open: b.Open, High: b.High,
			Low: b.Low, Close: b.Close, Volume: b.Volume,
			AdjClose: b.AdjClose, Source: b.Source,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":   symbol,
		"interval": interval,
		"count":    len(out),
		"bars":     out,
	})
}

// GET /api/v1/prices/latest?symbol=AAPL&interval=1d
func (s *Server) getLatestPrice(c *gin.Context) {
	symbol := c.Query("symbol")
	interval := c.DefaultQuery("interval", "1d")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	bar, err := s.db.LatestBar(c.Request.Context(), symbol, interval)
	if err == sql.ErrNoRows || bar == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no data found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":    bar.Symbol,
		"interval":  bar.Interval,
		"ts":        bar.TS,
		"open":      bar.Open,
		"high":      bar.High,
		"low":       bar.Low,
		"close":     bar.Close,
		"volume":    bar.Volume,
		"adj_close": bar.AdjClose,
		"source":    bar.Source,
	})
}

// POST /api/v1/ingest  -- triggers a manual ingest run
func (s *Server) triggerIngest(c *gin.Context) {
	s.scheduler.RunNow(c.Request.Context())
	c.JSON(http.StatusAccepted, gin.H{"status": "ingest triggered"})
}

func parseDateRange(c *gin.Context) (from, to time.Time, err error) {
	toStr := c.DefaultQuery("to", time.Now().Format("2006-01-02"))
	fromStr := c.DefaultQuery("from", time.Now().AddDate(-1, 0, 0).Format("2006-01-02"))

	from, err = time.Parse("2006-01-02", fromStr)
	if err != nil {
		return from, to, err
	}
	to, err = time.Parse("2006-01-02", toStr)
	return from, to, err
}
