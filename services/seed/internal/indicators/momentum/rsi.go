package momentum

import (
	"fmt"
	"math"

	"github.com/finance/seed/internal/indicators"
)

func init() { indicators.Register(&RSI{}) }

// RSI - Relative Strength Index (Wilder's smoothing method)
// Params: period (int, default 14)
// Returns values in [0, 100]. NaN for the first period-1 bars.
type RSI struct{}

func (r *RSI) Name() string { return "RSI" }

func (r *RSI) Compute(close indicators.Series, p indicators.Params) (indicators.Result, error) {
	period := indicators.IntParam(p, "period", 14)
	if period < 1 {
		return indicators.Result{}, fmt.Errorf("RSI: period must be >= 1")
	}

	n := len(close)
	out := make(indicators.Series, n)
	for i := range out {
		out[i] = math.NaN()
	}
	if n < period+1 {
		return indicators.Result{Values: out}, nil
	}

	// Compute price changes
	gains := make([]float64, n)
	losses := make([]float64, n)
	for i := 1; i < n; i++ {
		diff := close[i] - close[i-1]
		if diff > 0 {
			gains[i] = diff
		} else {
			losses[i] = -diff
		}
	}

	// Seed: simple average of first `period` gains/losses
	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	out[period] = rsiValue(avgGain, avgLoss)

	// Wilder's smoothing for remaining bars
	for i := period + 1; i < n; i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)
		out[i] = rsiValue(avgGain, avgLoss)
	}

	return indicators.Result{Values: out}, nil
}

func rsiValue(avgGain, avgLoss float64) float64 {
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}
