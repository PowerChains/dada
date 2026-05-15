-- PostgreSQL schema for DADA aggregator

CREATE TABLE IF NOT EXISTS users (
    wallet TEXT PRIMARY KEY,
    prefs JSONB DEFAULT '{}'::jsonb,
    slippage_tolerance DOUBLE PRECISION DEFAULT 0.5
);

CREATE TABLE IF NOT EXISTS trades (
    id SERIAL PRIMARY KEY,
    wallet TEXT NOT NULL,
    token_in TEXT NOT NULL,
    token_out TEXT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    route JSONB NOT NULL,
    fees DOUBLE PRECISION NOT NULL,
    slippage DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS dex_data (
    pool_id TEXT PRIMARY KEY,
    dex TEXT NOT NULL,
    token_pair TEXT NOT NULL,
    liquidity DOUBLE PRECISION NOT NULL,
    fee DOUBLE PRECISION NOT NULL,
    slippage DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS arbitrage_opportunities (
    id SERIAL PRIMARY KEY,
    pair TEXT NOT NULL,
    dex_a TEXT NOT NULL,
    dex_b TEXT NOT NULL,
    price_diff DOUBLE PRECISION NOT NULL,
    required_capital DOUBLE PRECISION NOT NULL,
    est_profit DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
