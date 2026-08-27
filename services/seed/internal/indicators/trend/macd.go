package trend

import (
	"fmt"
	"math"

	"github.com/finance/seed/internal/indicators"
)

func init() { indicators.Register(&MACD{}) }

// MACD - Moving Average Convergence/Divergence
// Params:
//   fast   (int, default 12)
//   slow   (int, default 26)
//   signal (int, default 9)
//
// Result:
//   Values         = MACD line  (EMA_fast - EMA_slow)
//   Extra["signal"]    = Signal line (EMA_signal of MACD)
//   Extra["histogram"] = MACD - Signal
type MACD struct{}

func (m *MACD) Name() string { return "MACD" }

func (m *MACD) Compute(close indicators.Series, p indicators.Params) (indicators.Result, error) {
	fast := indicators.IntParam(p, "fast", 12)
	slow := indicators.IntParam(p, "slow", 26)
	sig := indicators.IntParam(p, "signal", 9)

	if fast >= slow {
		return indicators.Result{}, fmt.Errorf("MACD: fast (%d) must be < slow (%d)", fast, slow)
	}

	n := len(close)
	emaFast := EMASlice(close, fast)
	emaSlow := EMASlice(close, slow)

	macdLine := make(indicators.Series, n)
	for i := range macdLine {
		if math.IsNaN(emaFast[i]) || math.IsNaN(emaSlow[i]) {
			macdLine[i] = math.NaN()
		} else {
			macdLine[i] = emaFast[i] - emaSlow[i]
		}
	}

	signalLine := EMASlice(macdLine, sig)

	histogram := make(indicators.Series, n)
	for i := range histogram {
		if math.IsNaN(macdLine[i]) || math.IsNaN(signalLine[i]) {
			histogram[i] = math.NaN()
		} else {
			histogram[i] = macdLine[i] - signalLine[i]
		}
	}

	return indicators.Result{
		Values: macdLine,
		Extra: map[string]indicators.Series{
			"signal":    signalLine,
			"histogram": histogram,
		},
	}, nil
}
