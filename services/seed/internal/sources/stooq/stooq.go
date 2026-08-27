package stooq

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/finance/seed/internal/sources"
)

// Stooq provides free historical EOD data (no API key required).
// Docs: https://stooq.com/
//
// Supported intervals: daily (1d), weekly (1wk), monthly (1mo).
// Symbol format: AAPL.US, MSFT.US, ^SPX (indices), BTCUSD.V (crypto).

const baseURL = "https://stooq.com/q/d/l/"

var intervalCode = map[string]string{
	"1d":  "d",
	"1wk": "w",
	"1mo": "m",
}

// Source implements sources.Source for Stooq.
type Source struct {
	client *http.Client
}

func New(timeout time.Duration) *Source {
	return &Source{
		client: &http.Client{Timeout: timeout},
	}
}

func (s *Source) Name() string { return "stooq" }

func (s *Source) SupportedIntervals() []string {
	return []string{"1d", "1wk", "1mo"}
}

func (s *Source) Fetch(ctx context.Context, symbol, interval string, from, to time.Time) ([]sources.Bar, error) {
	code, ok := intervalCode[interval]
	if !ok {
		return nil, fmt.Errorf("stooq does not support interval %q", interval)
	}

	stooqSymbol := toStooqSymbol(symbol)
	url := fmt.Sprintf("%s?s=%s&d1=%s&d2=%s&i=%s",
		baseURL,
		stooqSymbol,
		from.Format("20060102"),
		to.Format("20060102"),
		code,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; finance-seed/1.0)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch stooq %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stooq returned HTTP %d for %s", resp.StatusCode, symbol)
	}

	return parseCSV(symbol, interval, resp)
}

func parseCSV(symbol, interval string, resp *http.Response) ([]sources.Bar, error) {
	r := csv.NewReader(resp.Body)
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("no data rows for %s", symbol)
	}

	// Stooq returns newest-first; we reverse to oldest-first.
	bars := make([]sources.Bar, 0, len(records)-1)
	for i := len(records) - 1; i >= 1; i-- {
		row := records[i]
		if len(row) < 5 {
			continue
		}

		ts, err := time.Parse("2006-01-02", row[0])
		if err != nil {
			continue
		}

		open, err1 := strconv.ParseFloat(row[1], 64)
		high, err2 := strconv.ParseFloat(row[2], 64)
		low, err3 := strconv.ParseFloat(row[3], 64)
		close_, err4 := strconv.ParseFloat(row[4], 64)
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			continue
		}

		bar := sources.Bar{
			Symbol:   symbol,
			Interval: interval,
			TS:       ts.UTC(),
			Open:     open,
			High:     high,
			Low:      low,
			Close:    close_,
			Source:   "stooq",
		}

		if len(row) >= 6 {
			if vol, err := strconv.ParseInt(row[5], 10, 64); err == nil {
				bar.Volume = vol
			}
		}

		bars = append(bars, bar)
	}

	return bars, nil
}

// toStooqSymbol converts a canonical symbol to Stooq format.
// Yahoo-style crypto "BTC-USD" -> "BTCUSD.V", stocks "AAPL" -> "AAPL.US"
func toStooqSymbol(symbol string) string {
	upper := strings.ToUpper(symbol)

	// Crypto: BTC-USD -> BTCUSD.V
	if strings.HasSuffix(upper, "-USD") {
		base := strings.TrimSuffix(upper, "-USD")
		return strings.ToLower(base+"USD") + ".v"
	}

	// Indices already in Stooq format (e.g. ^SPX -> ^spx)
	if strings.HasPrefix(symbol, "^") {
		return strings.ToLower(symbol)
	}

	// US stocks: AAPL -> aapl.us
	return strings.ToLower(upper) + ".us"
}
