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

---

## Running Locally (Kubernetes)

The full stack runs on a local [kind](https://kind.sigs.k8s.io/) cluster managed by Terraform.
Services are exposed via Envoy on `localhost:8080`.

### Prerequisites

| Tool      | Version  | Install                              |
|-----------|----------|--------------------------------------|
| Docker    | any      | https://docs.docker.com/get-docker   |
| kind      | >= 0.23  | `brew install kind`                  |
| Terraform | >= 1.6   | `brew install terraform`             |
| kubectl   | >= 1.30  | `brew install kubectl`               |
| helm      | >= 3.12  | `brew install helm`                  |

### 1. Provision the cluster

```bash
cd infrastructure/terraform/environments/local

terraform init
terraform apply
```

This creates a 3-node kind cluster (`finance-local`) and the `finance` namespace.

### 2. Configure kubectl

```bash
kind get kubeconfig --name finance-local > ~/.kube/config-finance-local
export KUBECONFIG=~/.kube/config-finance-local

kubectl get nodes   # should show 3 nodes Ready
```

### 3. Build and load service images

Each service image must be built locally and loaded into kind (kind clusters cannot pull from Docker Hub or ECR by default).

```bash
# Seed service (the only service currently implemented)
docker build -t finance/seed:local services/seed/
kind load docker-image finance/seed:local --name finance-local

# Repeat for other services as they are implemented:
# docker build -t finance/analytics:local services/analytics/
# docker build -t finance/alert:local     services/alert/
# kind load docker-image finance/analytics:local --name finance-local
# kind load docker-image finance/alert:local     --name finance-local
```

### 4. Deploy services

```bash
kubectl apply -k infrastructure/k8s/overlays/local/
```

Verify everything is running:

```bash
kubectl get pods -n finance
kubectl get svc  -n finance
```

### 5. Deploy observability stack (optional)

```bash
cd infrastructure/k8s/monitoring
chmod +x deploy.sh && ./deploy.sh
```

Access Grafana (metrics, logs, and traces in one UI):

```bash
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
open http://localhost:3000   # admin / changeme
```

### 6. Call the API

All traffic goes through Envoy on `localhost:8080`.

```bash
# Health check
curl http://localhost:8080/health

# Latest price
curl "http://localhost:8080/api/v1/prices/latest?symbol=AAPL&interval=1d"

# Historical prices
curl "http://localhost:8080/api/v1/prices?symbol=AAPL&interval=1d&from=2024-01-01&to=2024-12-31"

# Indicator
curl "http://localhost:8080/api/v1/indicators?symbol=AAPL&interval=1d&indicator=RSI&period=14"

# Trigger manual ingest
curl -X POST http://localhost:8080/api/v1/ingest
```

### Tear down

```bash
# Remove all k8s resources
kubectl delete -k infrastructure/k8s/overlays/local/

# Destroy the kind cluster
cd infrastructure/terraform/environments/local
terraform destroy
```
