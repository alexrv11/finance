# Finance

A financial data platform for ingesting, storing, and analyzing stock and crypto market data.

## Services

### `seed`

A Go microservice that fetches OHLCV (Open, High, Low, Close, Volume) data from free public sources, stores it in DuckDB, computes technical indicators, and exposes a REST API.

**Stack:** Go 1.22 · DuckDB · Gin · cron

---

## Getting Started

### Prerequisites

- Go 1.22+
- No API keys required for default sources (Yahoo Finance, Stooq)

### Setup

```bash
cd services/seed

# Create data directory
make data-dir

# Copy and configure environment
cp .env.example .env

# Build
make build

# Run
make run
```

### Environment Variables

| Variable             | Default                     | Description                          |
|----------------------|-----------------------------|--------------------------------------|
| `DB_PATH`            | `./data/finance.duckdb`     | DuckDB database file path            |
| `API_PORT`           | `8080`                      | HTTP server port                     |
| `SYMBOLS`            | `AAPL,MSFT,...,BTC-USD,...` | Comma-separated symbols to track     |
| `SCHEDULE_INTERVAL`  | `0 * * * *`                 | Cron expression for data ingestion   |
| `FETCH_TIMEOUT_SECS` | `30`                        | HTTP timeout for source fetches      |
| `ALPHA_VANTAGE_KEY`  | —                           | Optional — Alpha Vantage API key     |
| `FMP_KEY`            | —                           | Optional — Financial Modeling Prep   |
| `FRED_KEY`           | —                           | Optional — FRED API key              |

---

## API

Base path: `/api/v1`

### Prices

```
GET /api/v1/prices?symbol=AAPL&interval=1d&from=2024-01-01&to=2024-12-31
GET /api/v1/prices/latest?symbol=AAPL&interval=1d
POST /api/v1/ingest
```

### Indicators

```
GET /api/v1/indicators?symbol=AAPL&interval=1d&indicator=RSI&period=14
```

Supported indicators: `SMA`, `EMA`, `MACD`, `RSI`, `BB` (Bollinger Bands)

### Health

```
GET /health
```

---

## Data Sources

| Source        | Intervals          | Notes                        |
|---------------|--------------------|------------------------------|
| Yahoo Finance | 1m, 5m, 1h, 1d, … | Stocks, ETFs, crypto, no key |
| Stooq         | 1d, 1wk, 1mo       | Global EOD data, no key      |

Ingestion runs on schedule (default: hourly) and can be triggered manually via `POST /api/v1/ingest`.

---

## Project Structure

```
services/seed/
  cmd/seed/main.go                      # Entrypoint
  internal/
    config/config.go                    # Env-var config
    api/
      server.go                         # Gin server + routes
      prices.go                         # Price endpoints
      indicators_handler.go             # Indicator endpoint
    ingest/
      scheduler.go                      # Cron-based ingestion
      normalizer.go                     # Dedup, filter, sort
    sources/
      source.go                         # Source interface + Bar struct
      yahoo/yahoo.go                    # Yahoo Finance adapter
      stooq/stooq.go                    # Stooq adapter
    store/
      db.go                             # DuckDB init + migrations
      schema.sql                        # DDL
      ohlcv.go                          # Read/write OHLCV
    indicators/
      indicator.go                      # Interface + registry
      trend/                            # SMA, EMA, MACD
      momentum/                         # RSI
      volatility/                       # Bollinger Bands
```

---

## Development

```bash
make build   # Compile binary to bin/seed
make run     # Run without building
make test    # Run all tests
make tidy    # go mod tidy
make lint    # golangci-lint (requires golangci-lint installed)
```
