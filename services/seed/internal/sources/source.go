package sources

import (
	"context"
	"time"
)

// Bar represents a single OHLCV price bar.
type Bar struct {
	Symbol   string
	Interval string
	TS       time.Time
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   int64
	AdjClose float64
	Source   string
}

// Source is the interface all data source adapters must implement.
type Source interface {
	// Name returns the unique identifier for this source (e.g. "yahoo", "stooq").
	Name() string

	// Fetch retrieves OHLCV bars for the given symbol and interval in [from, to].
	Fetch(ctx context.Context, symbol, interval string, from, to time.Time) ([]Bar, error)

	// SupportedIntervals lists the interval strings this source can provide.
	SupportedIntervals() []string
}

// IntervalDuration maps canonical interval strings to their duration.
var IntervalDuration = map[string]time.Duration{
	"1m":  time.Minute,
	"2m":  2 * time.Minute,
	"5m":  5 * time.Minute,
	"15m": 15 * time.Minute,
	"30m": 30 * time.Minute,
	"1h":  time.Hour,
	"4h":  4 * time.Hour,
	"1d":  24 * time.Hour,
	"1wk": 7 * 24 * time.Hour,
	"1mo": 30 * 24 * time.Hour,
}
