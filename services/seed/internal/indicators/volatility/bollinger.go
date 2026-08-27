package volatility

import (
	"fmt"
	"math"

	"github.com/finance/seed/internal/indicators"
	"github.com/finance/seed/internal/indicators/trend"
)

func init() { indicators.Register(&BollingerBands{}) }

// BollingerBands computes upper, middle, and lower Bollinger Bands.
// Params:
//   period (int,     default 20)
//   stddev (float64, default 2.0)
//
// Result:
//   Values          = middle band (SMA)
//   Extra["upper"]  = upper band
//   Extra["lower"]  = lower band
//   Extra["width"]  = (upper - lower) / middle  (bandwidth)
//   Extra["pct_b"]  = (close - lower) / (upper - lower) (%B)
type BollingerBands struct{}

func (b *BollingerBands) Name() string { return "BBANDS" }

func (b *BollingerBands) Compute(close indicators.Series, p indicators.Params) (indicators.Result, error) {
	period := indicators.IntParam(p, "period", 20)
	mult := indicators.FloatParam(p, "stddev", 2.0)

	if period < 2 {
		return indicators.Result{}, fmt.Errorf("BBANDS: period must be >= 2")
	}

	n := len(close)
	middle := trend.SMASlice(close, period)
	upper := make(indicators.Series, n)
	lower := make(indicators.Series, n)
	width := make(indicators.Series, n)
	pctB := make(indicators.Series, n)

	for i := range upper {
		upper[i] = math.NaN()
		lower[i] = math.NaN()
		width[i] = math.NaN()
		pctB[i] = math.NaN()
	}

	for i := period - 1; i < n; i++ {
		if math.IsNaN(middle[i]) {
			continue
		}
		sd := stdDev(close[i-period+1:i+1], middle[i])
		upper[i] = middle[i] + mult*sd
		lower[i] = middle[i] - mult*sd

		bw := upper[i] - lower[i]
		if middle[i] != 0 {
			width[i] = bw / middle[i]
		}
		if bw != 0 {
			pctB[i] = (close[i] - lower[i]) / bw
		}
	}

	return indicators.Result{
		Values: middle,
		Extra: map[string]indicators.Series{
			"upper": upper,
			"lower": lower,
			"width": width,
			"pct_b": pctB,
		},
	}, nil
}

func stdDev(data indicators.Series, mean float64) float64 {
	var sum float64
	for _, v := range data {
		d := v - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(data)))
}
