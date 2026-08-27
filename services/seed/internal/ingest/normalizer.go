package ingest

import (
	"sort"
	"time"

	"github.com/finance/seed/internal/sources"
)

// Deduplicate removes bars with duplicate (symbol, interval, ts) keeping
// the first occurrence (highest priority source should be passed first).
func Deduplicate(bars []sources.Bar) []sources.Bar {
	seen := make(map[string]struct{}, len(bars))
	out := make([]sources.Bar, 0, len(bars))
	for _, b := range bars {
		key := b.Symbol + "|" + b.Interval + "|" + b.TS.Format(time.RFC3339)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, b)
	}
	return out
}

// SortByTS sorts bars ascending by timestamp.
func SortByTS(bars []sources.Bar) {
	sort.Slice(bars, func(i, j int) bool {
		return bars[i].TS.Before(bars[j].TS)
	})
}

// FilterValid drops bars where OHLC values are non-positive or close == 0.
func FilterValid(bars []sources.Bar) []sources.Bar {
	out := make([]sources.Bar, 0, len(bars))
	for _, b := range bars {
		if b.Open <= 0 || b.High <= 0 || b.Low <= 0 || b.Close <= 0 {
			continue
		}
		if b.High < b.Low {
			continue
		}
		out = append(out, b)
	}
	return out
}

// Normalize runs deduplication, validity filtering, and sorting in one pass.
func Normalize(bars []sources.Bar) []sources.Bar {
	bars = FilterValid(bars)
	bars = Deduplicate(bars)
	SortByTS(bars)
	return bars
}
