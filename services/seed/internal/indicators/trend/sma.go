package trend

import (
	"fmt"
	"math"

	"github.com/finance/seed/internal/indicators"
)

func init() { indicators.Register(&SMA{}) }

// SMA - Simple Moving Average
// Params: period (int, default 20)
type SMA struct{}

func (s *SMA) Name() string { return "SMA" }

func (s *SMA) Compute(close indicators.Series, p indicators.Params) (indicators.Result, error) {
	period := indicators.IntParam(p, "period", 20)
	if period < 1 {
		return indicators.Result{}, fmt.Errorf("SMA: period must be >= 1")
	}
	n := len(close)
	out := make(indicators.Series, n)
	for i := range out {
		out[i] = math.NaN()
	}

	for i := period - 1; i < n; i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += close[j]
		}
		out[i] = sum / float64(period)
	}

	return indicators.Result{Values: out}, nil
}

// SMASlice is a convenience function used by other indicators internally.
func SMASlice(data indicators.Series, period int) indicators.Series {
	ind := &SMA{}
	res, _ := ind.Compute(data, indicators.Params{"period": period})
	return res.Values
}
