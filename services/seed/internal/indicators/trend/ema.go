package trend

import (
	"fmt"
	"math"

	"github.com/finance/seed/internal/indicators"
)

func init() { indicators.Register(&EMA{}) }

// EMA - Exponential Moving Average
// Params: period (int, default 20)
// Uses standard multiplier: k = 2 / (period + 1)
type EMA struct{}

func (e *EMA) Name() string { return "EMA" }

func (e *EMA) Compute(close indicators.Series, p indicators.Params) (indicators.Result, error) {
	period := indicators.IntParam(p, "period", 20)
	if period < 1 {
		return indicators.Result{}, fmt.Errorf("EMA: period must be >= 1")
	}
	return indicators.Result{Values: EMASlice(close, period)}, nil
}

// EMASlice computes EMA and is used internally by MACD, DEMA, etc.
// Seeds with the first SMA then applies the EMA formula.
func EMASlice(data indicators.Series, period int) indicators.Series {
	n := len(data)
	out := make(indicators.Series, n)
	for i := range out {
		out[i] = math.NaN()
	}
	if n < period {
		return out
	}

	k := 2.0 / float64(period+1)

	// Seed: first EMA value = SMA of first `period` bars
	seed := 0.0
	for i := 0; i < period; i++ {
		seed += data[i]
	}
	out[period-1] = seed / float64(period)

	for i := period; i < n; i++ {
		out[i] = data[i]*k + out[i-1]*(1-k)
	}
	return out
}
