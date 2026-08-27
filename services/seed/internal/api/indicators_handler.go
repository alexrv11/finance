package api

import (
	"math"
	"net/http"
	"time"

	"github.com/finance/seed/internal/indicators"
	"github.com/gin-gonic/gin"

	// blank imports trigger init() registration for each indicator package
	_ "github.com/finance/seed/internal/indicators/momentum"
	_ "github.com/finance/seed/internal/indicators/trend"
	_ "github.com/finance/seed/internal/indicators/volatility"
)

// GET /api/v1/indicators?symbol=AAPL&interval=1d&indicator=RSI&period=14&from=2024-01-01
func (s *Server) getIndicator(c *gin.Context) {
	symbol := c.Query("symbol")
	interval := c.DefaultQuery("interval", "1d")
	indName := c.Query("indicator")

	if symbol == "" || indName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol and indicator are required"})
		return
	}

	from, to, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch a larger window to warm up the indicator (extra 200 bars).
	// We'll trim the output back to the requested [from, to] range.
	warmup := 200
	extFrom := from.AddDate(-1, 0, 0)

	bars, err := s.db.QueryBars(c.Request.Context(), symbol, interval, extFrom, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(bars) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no price data found"})
		return
	}

	// Build close series.
	close := make(indicators.Series, len(bars))
	for i, b := range bars {
		close[i] = b.Close
	}

	// Build params from query string.
	params := indicators.Params{}
	for k, v := range c.Request.URL.Query() {
		if k == "symbol" || k == "interval" || k == "indicator" || k == "from" || k == "to" {
			continue
		}
		params[k] = v[0]
	}

	result, err := indicators.Compute(indName, close, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build response, trimming to requested range.
	type point struct {
		TS    time.Time          `json:"ts"`
		Value float64            `json:"value"`
		Extra map[string]float64 `json:"extra,omitempty"`
	}

	_ = warmup
	var points []point
	for i, b := range bars {
		if b.TS.Before(from) || b.TS.After(to) {
			continue
		}
		v := result.Values[i]
		if math.IsNaN(v) {
			continue
		}
		pt := point{TS: b.TS, Value: v}
		if len(result.Extra) > 0 {
			pt.Extra = make(map[string]float64)
			for k, series := range result.Extra {
				if i < len(series) && !math.IsNaN(series[i]) {
					pt.Extra[k] = series[i]
				}
			}
		}
		points = append(points, pt)
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":    symbol,
		"interval":  interval,
		"indicator": indName,
		"params":    params,
		"count":     len(points),
		"data":      points,
	})
}
