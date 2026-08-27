-- Symbols registry
CREATE TABLE IF NOT EXISTS symbols (
    symbol      VARCHAR PRIMARY KEY,
    name        VARCHAR,
    exchange    VARCHAR,
    asset_type  VARCHAR,   -- 'stock', 'etf', 'crypto', 'fx', 'index'
    currency    VARCHAR DEFAULT 'USD',
    active      BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- Canonical OHLCV price bars
CREATE TABLE IF NOT EXISTS ohlcv (
    symbol      VARCHAR NOT NULL,
    source      VARCHAR NOT NULL,   -- 'yahoo', 'stooq', 'alphavantage', etc.
    interval    VARCHAR NOT NULL,   -- '1m', '5m', '1h', '1d', '1wk', '1mo'
    ts          TIMESTAMPTZ NOT NULL,
    open        DOUBLE NOT NULL,
    high        DOUBLE NOT NULL,
    low         DOUBLE NOT NULL,
    close       DOUBLE NOT NULL,
    volume      BIGINT,
    adj_close   DOUBLE,
    PRIMARY KEY (symbol, source, interval, ts)
);

-- Computed indicator cache
CREATE TABLE IF NOT EXISTS indicator_cache (
    symbol      VARCHAR NOT NULL,
    interval    VARCHAR NOT NULL,
    indicator   VARCHAR NOT NULL,   -- e.g. 'RSI_14', 'EMA_20', 'BBANDS_20_2'
    ts          TIMESTAMPTZ NOT NULL,
    value       DOUBLE NOT NULL,
    extra       JSON,               -- for multi-value indicators (bands, MACD)
    PRIMARY KEY (symbol, interval, indicator, ts)
);

-- Ingest run log
CREATE TABLE IF NOT EXISTS ingest_log (
    id          VARCHAR PRIMARY KEY,
    symbol      VARCHAR NOT NULL,
    source      VARCHAR NOT NULL,
    interval    VARCHAR NOT NULL,
    started_at  TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    bars_saved  INTEGER DEFAULT 0,
    status      VARCHAR DEFAULT 'running',  -- 'running', 'ok', 'error'
    error_msg   VARCHAR
);
