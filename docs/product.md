# Product Design

## Vision

A self-hosted financial data platform that gives developers and traders programmatic access to clean market data and technical analysis — with no vendor lock-in and no per-call fees.

---

## Target Users

| Persona              | Description                                                                 |
|----------------------|-----------------------------------------------------------------------------|
| Quant / Developer    | Needs reliable OHLCV data and indicators via API to build trading strategies |
| Active Trader        | Monitors watchlists, sets price/indicator alerts, reviews charts            |
| Data Analyst         | Exports historical data for research, backtesting, and modeling             |

---

## Core Features

### 1. Market Data Ingestion
- Fetch OHLCV data for stocks, ETFs, indices, and crypto
- Support multiple free data sources with automatic fallback
- Scheduled ingestion (hourly by default) + on-demand trigger
- Supported intervals: `1m`, `5m`, `15m`, `1h`, `1d`, `1wk`, `1mo`

### 2. Technical Indicators
- Compute indicators on stored price data
- Trend: SMA, EMA, MACD
- Momentum: RSI, Stochastic, CCI *(planned)*
- Volatility: Bollinger Bands
- Volume: OBV, VWAP *(planned)*

### 3. Alerts *(planned)*
- Watchlist management per user
- Alert rules: price threshold, indicator crossover, percentage move
- Delivery channels: webhook, email

### 4. REST API
- Query prices and indicators over HTTP
- Consistent response format with pagination
- OpenAPI spec for easy client generation

### 5. gRPC API *(planned)*
- Low-latency inter-service and external access
- Streaming support for real-time price ticks

---

## User Flows

### Fetch Latest Price

```
User calls GET /api/v1/prices/latest?symbol=AAPL&interval=1d
    → Returns most recent OHLCV bar for AAPL
```

### Query Historical Data

```
User calls GET /api/v1/prices?symbol=BTC-USD&interval=1h&from=2024-01-01&to=2024-06-30
    → Returns paginated OHLCV bars for the requested range
```

### Compute an Indicator

```
User calls GET /api/v1/indicators?symbol=AAPL&interval=1d&indicator=RSI&period=14
    → Returns RSI(14) values aligned to each price bar
```

### Trigger Manual Ingest

```
User calls POST /api/v1/ingest  { "symbols": ["TSLA", "SPY"] }
    → Immediately fetches latest data for specified symbols
    → Returns ingest summary (bars fetched, errors)
```

---

## API Design Principles

- **RESTful:** Resources are nouns, HTTP verbs carry intent
- **Consistent errors:** All errors return `{ "error": "<message>", "code": "<code>" }`
- **Versioned:** All endpoints under `/api/v1/`
- **No auth required** for local/self-hosted deployment by default
- **Optional JWT auth** when exposed to external clients via API Gateway

---

## Data Model

### Symbol
```
symbol      string    e.g. "AAPL", "BTC-USD", "^SPX"
source      string    e.g. "yahoo", "stooq"
asset_type  string    stock | etf | crypto | index
```

### Price Bar (OHLCV)
```
symbol      string
interval    string    1m | 5m | 1h | 1d | 1wk | 1mo
timestamp   datetime
open        float64
high        float64
low         float64
close       float64
volume      float64
```

### Indicator Result
```
symbol      string
interval    string
indicator   string    RSI | SMA | EMA | MACD | BB
params      map       e.g. { "period": 14 }
timestamp   datetime
value       float64 | object   (MACD and BB return composite values)
```

---

## Roadmap

### Phase 1 — Foundation *(current)*
- [x] OHLCV ingestion from Yahoo Finance and Stooq
- [x] DuckDB storage with schema migrations
- [x] REST API: prices + indicators
- [x] Cron-based scheduling

### Phase 2 — Microservices
- [ ] Extract Analytics Service (indicators) as standalone gRPC service
- [ ] Add API Gateway with auth and routing
- [ ] Hexagonal architecture refactor of Seed Service
- [ ] Shared proto definitions

### Phase 3 — Alerts & Streaming
- [ ] Alert Service with watchlist management
- [ ] gRPC streaming for real-time price ticks
- [ ] Webhook + email notification delivery

### Phase 4 — Observability & Scale
- [ ] Structured logging (JSON) across all services
- [ ] Prometheus metrics + Grafana dashboards
- [ ] Docker Compose for local multi-service setup
- [ ] Kubernetes manifests for production deployment
