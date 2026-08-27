package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DBPath           string
	APIPort          string
	AlphaVantageKey  string
	FMPKey           string
	FREDKey          string
	FetchTimeout     time.Duration
	ScheduleInterval string
	Symbols          []string
}

func Load() Config {
	symbols := []string{}
	if raw := getEnv("SYMBOLS", "AAPL,MSFT,GOOGL,SPY,BTC-USD"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			if t := strings.TrimSpace(s); t != "" {
				symbols = append(symbols, t)
			}
		}
	}

	return Config{
		DBPath:           getEnv("DB_PATH", "./data/finance.duckdb"),
		APIPort:          getEnv("API_PORT", "8080"),
		AlphaVantageKey:  getEnv("ALPHA_VANTAGE_KEY", ""),
		FMPKey:           getEnv("FMP_KEY", ""),
		FREDKey:          getEnv("FRED_KEY", ""),
		FetchTimeout:     getDuration("FETCH_TIMEOUT_SECS", 30),
		ScheduleInterval: getEnv("SCHEDULE_INTERVAL", "0 * * * *"),
		Symbols:          symbols,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallbackSecs int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Second
		}
	}
	return time.Duration(fallbackSecs) * time.Second
}
