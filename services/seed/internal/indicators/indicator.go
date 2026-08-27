package indicators

import (
	"fmt"
	"math"
	"strconv"
)

// Series is a slice of float64 values aligned to a time series.
// NaN values represent missing/insufficient data (head padding).
type Series []float64

// Params holds named parameters for indicator computation.
type Params map[string]any

// Result holds the output of a (possibly multi-value) indicator.
type Result struct {
	// Values is the primary series (e.g. MACD line, BB middle, RSI).
	Values Series
	// Extra holds named supplementary series (e.g. "signal", "upper", "lower").
	Extra map[string]Series
}

// Indicator computes a technical indicator over a price series.
type Indicator interface {
	Name() string
	Compute(close Series, params Params) (Result, error)
}

// Registry holds all registered indicators keyed by name.
var Registry = map[string]Indicator{}

// Register adds an indicator to the global registry.
func Register(ind Indicator) {
	Registry[ind.Name()] = ind
}

// Compute looks up an indicator by name and runs it.
func Compute(name string, close Series, params Params) (Result, error) {
	ind, ok := Registry[name]
	if !ok {
		return Result{}, fmt.Errorf("unknown indicator %q", name)
	}
	return ind.Compute(close, params)
}

// --- Exported param helpers -------------------------------------------------
// These handle int, float64, and string (from query params) gracefully.

func IntParam(p Params, key string, def int) int {
	v, ok := p[key]
	if !ok {
		return def
	}
	switch t := v.(type) {
	case int:
		return t
	case float64:
		return int(t)
	case string:
		if n, err := strconv.Atoi(t); err == nil {
			return n
		}
	}
	return def
}

func FloatParam(p Params, key string, def float64) float64 {
	v, ok := p[key]
	if !ok {
		return def
	}
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case string:
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			return f
		}
	}
	return def
}

func NaNSeries(n int) Series {
	s := make(Series, n)
	for i := range s {
		s[i] = math.NaN()
	}
	return s
}
