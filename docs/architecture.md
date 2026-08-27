# Architecture Design

## Overview

The platform is built as a **microservice architecture** where services communicate internally via **gRPC** and expose a unified surface to external clients through an **API Gateway** over HTTP/REST.

Each service follows **Hexagonal Architecture** (Ports & Adapters), keeping business logic isolated from infrastructure concerns.

---

## System Diagram

```
                        ┌─────────────────────────────────────────────────────┐
                        │                   External Clients                  │
                        │           (Web App, Mobile, Third-party)            │
                        └───────────────────────┬─────────────────────────────┘
                                                │ HTTP / REST
                        ┌───────────────────────▼─────────────────────────────┐
                        │                   API Gateway                       │
                        │         Auth · Rate Limiting · Routing              │
                        └───┬───────────────┬───────────────┬─────────────────┘
                            │ gRPC          │ gRPC          │ gRPC
              ┌─────────────▼──┐   ┌────────▼───────┐  ┌───▼────────────┐
              │  Seed Service  │   │ Analytics Svc  │  │  Alert Service │
              │  (market data) │   │  (indicators)  │  │  (watchlists)  │
              └───────┬────────┘   └────────┬───────┘  └───────┬────────┘
                      │                     │                   │
              ┌───────▼────────┐   ┌────────▼───────┐  ┌───────▼────────┐
              │    DuckDB      │   │   TimescaleDB  │  │   PostgreSQL   │
              └────────────────┘   └────────────────┘  └────────────────┘
```

---

## Services

### API Gateway
- Single entry point for all external traffic
- Responsibilities: JWT auth, rate limiting, request routing, TLS termination
- Communicates downstream via gRPC

### Seed Service *(exists)*
- Ingests OHLCV market data from external sources (Yahoo Finance, Stooq)
- Runs on a cron schedule or on-demand trigger
- Exposes gRPC endpoints for querying raw price data

### Analytics Service *(planned)*
- Computes technical indicators (SMA, EMA, MACD, RSI, Bollinger Bands)
- Pulls raw bars from Seed via gRPC
- Caches computed results

### Alert Service *(planned)*
- Manages user watchlists and price/indicator alerts
- Subscribes to market data events via gRPC streaming
- Dispatches notifications (email, webhook)

---

## Hexagonal Architecture

Every microservice is structured around the same hexagonal pattern:

```
┌───────────────────────────────────────────────────────┐
│                      Service                          │
│                                                       │
│  ┌─────────────────────────────────────────────────┐  │
│  │              PRIMARY ADAPTERS (Driving)         │  │
│  │   gRPC Handler · REST Handler · CLI · Scheduler │  │
│  └──────────────────────┬──────────────────────────┘  │
│                         │ calls                       │
│  ┌──────────────────────▼──────────────────────────┐  │
│  │                  PRIMARY PORTS                  │  │
│  │            (Inbound use-case interfaces)        │  │
│  └──────────────────────┬──────────────────────────┘  │
│                         │                             │
│  ┌──────────────────────▼──────────────────────────┐  │
│  │              APPLICATION CORE                   │  │
│  │         Domain models · Use cases               │  │
│  │         Business rules · Domain events          │  │
│  └──────────────────────┬──────────────────────────┘  │
│                         │                             │
│  ┌──────────────────────▼──────────────────────────┐  │
│  │                SECONDARY PORTS                  │  │
│  │           (Outbound repository interfaces)      │  │
│  └──────────────────────┬──────────────────────────┘  │
│                         │ implemented by              │
│  ┌──────────────────────▼──────────────────────────┐  │
│  │            SECONDARY ADAPTERS (Driven)          │  │
│  │    DuckDB · PostgreSQL · HTTP Client · gRPC     │  │
│  └─────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────┘
```

### Folder Structure per Service

```
services/<name>/
  cmd/<name>/
    main.go                 # Wire adapters → core → start server
  internal/
    domain/
      model.go              # Domain entities and value objects
      events.go             # Domain events
    ports/
      inbound.go            # Primary port interfaces (use-case contracts)
      outbound.go           # Secondary port interfaces (repo/client contracts)
    application/
      <use_case>.go         # Use-case implementations
    adapters/
      grpc/
        server.go           # gRPC handler (primary adapter)
        proto/              # .proto definitions + generated code
      http/
        server.go           # REST handler if needed (primary adapter)
      store/
        <db>_repo.go        # DB repository (secondary adapter)
      sources/
        <name>_client.go    # External HTTP/API clients (secondary adapter)
    config/
      config.go
```

---

## gRPC Communication

### Protocol

- All inter-service calls use **gRPC** over HTTP/2
- Proto files live in each service's `adapters/grpc/proto/` directory
- A shared `proto/` directory at the repo root holds cross-service contracts

### Shared Proto Layout

```
proto/
  seed/v1/
    prices.proto            # PriceBar, GetBarsRequest, GetBarsResponse
    ingest.proto            # TriggerIngestRequest, TriggerIngestResponse
  analytics/v1/
    indicators.proto        # IndicatorRequest, IndicatorResponse
  alert/v1/
    alerts.proto            # Alert, WatchlistRequest
```

### Example: Seed Service gRPC Contract

```protobuf
syntax = "proto3";
package seed.v1;

service PriceService {
  rpc GetBars (GetBarsRequest) returns (GetBarsResponse);
  rpc GetLatestBar (GetLatestBarRequest) returns (PriceBar);
  rpc StreamBars (GetBarsRequest) returns (stream PriceBar);
}

message PriceBar {
  string symbol    = 1;
  string interval  = 2;
  int64  timestamp = 3;
  double open      = 4;
  double high      = 5;
  double low       = 6;
  double close     = 7;
  double volume    = 8;
}

message GetBarsRequest {
  string symbol   = 1;
  string interval = 2;
  int64  from     = 3;
  int64  to       = 4;
}

message GetBarsResponse {
  repeated PriceBar bars = 1;
}
```

---

## Data Flow

### Ingest Flow

```
Scheduler / API trigger
    → Seed Application (use case: IngestBars)
        → Source Adapter (Yahoo / Stooq HTTP client)
        → Normalizer (dedup, sort, validate)
        → Store Adapter (DuckDB repo)
```

### Query Flow (external client)

```
Client HTTP request
    → API Gateway (auth, route)
        → gRPC call to Seed / Analytics
            → Application use case
                → Store or upstream gRPC
            → Response assembled
        → HTTP response to client
```

---

## Cross-Cutting Concerns

| Concern         | Approach                                      |
|-----------------|-----------------------------------------------|
| Authentication  | JWT validated at API Gateway only             |
| Observability   | Structured JSON logs · Prometheus metrics     |
| Error handling  | gRPC status codes + domain error types        |
| Configuration   | Env vars per service, no shared config store  |
| Testing         | Unit-test core with mock secondary ports      |
