package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/finance/seed/internal/sources"
)

const (
	baseURL   = "https://query1.finance.yahoo.com/v8/finance/chart"
	userAgent = "Mozilla/5.0 (compatible; finance-seed/1.0)"
)

// response mirrors the Yahoo Finance v8 chart API JSON shape.
type response struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol   string `json:"symbol"`
				Currency string `json:"currency"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []*float64 `json:"open"`
					High   []*float64 `json:"high"`
					Low    []*float64 `json:"low"`
					Close  []*float64 `json:"close"`
					Volume []*int64   `json:"volume"`
				} `json:"quote"`
				AdjClose []struct {
					AdjClose []*float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// Source implements sources.Source for Yahoo Finance (unofficial API).
type Source struct {
	client *http.Client
}

func New(timeout time.Duration) *Source {
	return &Source{
		client: &http.Client{Timeout: timeout},
	}
}

func (s *Source) Name() string { return "yahoo" }

func (s *Source) SupportedIntervals() []string {
	return []string{"1m", "2m", "5m", "15m", "30m", "60m", "1h", "1d", "5d", "1wk", "1mo", "3mo"}
}

func (s *Source) Fetch(ctx context.Context, symbol, interval string, from, to time.Time) ([]sources.Bar, error) {
	url := fmt.Sprintf("%s/%s?interval=%s&period1=%d&period2=%d&includeAdjustedClose=true",
		baseURL, symbol, interval, from.Unix(), to.Unix())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch yahoo %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo returned HTTP %d for %s", resp.StatusCode, symbol)
	}

	var yr response
	if err := json.NewDecoder(resp.Body).Decode(&yr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if yr.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo error %s: %s", yr.Chart.Error.Code, yr.Chart.Error.Description)
	}

	if len(yr.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned for %s", symbol)
	}

	return parseBars(symbol, interval, &yr.Chart.Result[0])
}

func parseBars(symbol, interval string, result *struct {
	Meta struct {
		Symbol   string `json:"symbol"`
		Currency string `json:"currency"`
	} `json:"meta"`
	Timestamp  []int64 `json:"timestamp"`
	Indicators struct {
		Quote []struct {
			Open   []*float64 `json:"open"`
			High   []*float64 `json:"high"`
			Low    []*float64 `json:"low"`
			Close  []*float64 `json:"close"`
			Volume []*int64   `json:"volume"`
		} `json:"quote"`
		AdjClose []struct {
			AdjClose []*float64 `json:"adjclose"`
		} `json:"adjclose"`
	} `json:"indicators"`
}) ([]sources.Bar, error) {
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data for %s", symbol)
	}

	q := result.Indicators.Quote[0]
	n := len(result.Timestamp)
	bars := make([]sources.Bar, 0, n)

	for i := 0; i < n; i++ {
		// Skip bars with nil OHLC (gaps in data)
		if q.Open[i] == nil || q.High[i] == nil || q.Low[i] == nil || q.Close[i] == nil {
			continue
		}

		bar := sources.Bar{
			Symbol:   symbol,
			Interval: interval,
			TS:       time.Unix(result.Timestamp[i], 0).UTC(),
			Open:     *q.Open[i],
			High:     *q.High[i],
			Low:      *q.Low[i],
			Close:    *q.Close[i],
			Source:   "yahoo",
		}

		if q.Volume[i] != nil {
			bar.Volume = *q.Volume[i]
		}

		// Adjusted close
		if len(result.Indicators.AdjClose) > 0 && i < len(result.Indicators.AdjClose[0].AdjClose) {
			if ac := result.Indicators.AdjClose[0].AdjClose[i]; ac != nil {
				bar.AdjClose = *ac
			}
		}

		if !math.IsNaN(bar.Open) && !math.IsNaN(bar.Close) {
			bars = append(bars, bar)
		}
	}

	return bars, nil
}
