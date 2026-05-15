# DADA App

DADA is a Solana multi-DEX aggregator with a Go API backend, Redis caching, PostgreSQL storage, and a React/Vite frontend.

## Features
- Best-price finder across 22+ Solana DEXs
- Swap simulator with route visualization
- Wallet connect support for Phantom, Solflare, Backpack
- AI suggestions and token search assistance
- Arbitrage monitoring with profit estimates and required capital
- Trade history analytics with wallet-aware retrieval

## Local development

1. Install dependencies:
   - `cd backend && go mod tidy`
   - `cd frontend && npm install`

2. Start the backend:
   - `cd backend && go run .`

3. Start the frontend:
   - `cd frontend && npm run dev`

4. Open the app:
   - `http://localhost:5173`

## Production-ready build

This repository includes production deployment artifacts:
- `backend/Dockerfile`
- `frontend/Dockerfile`
- `frontend/nginx.conf`
- `docker-compose.yml`
- `docs/production.md`

Build and run using Docker Compose:

```bash
make setup
make docker-build
make docker-up
```

Open the app at:

- `http://localhost:3000`

The backend API remains available on:

- `http://localhost:8080`

## Environment

Backend environment variables:
- `DATABASE_DSN` - Postgres connection string (default: `postgres://dada:password@db:5432/dada?sslmode=disable`)
- `REDIS_ADDR` - Redis host address (default: `redis:6379`)
- `PORT` - HTTP bind port (default: `8080`)
- `AISTUDIO_KEY` - Optional AI key for assistant features

Frontend environment variables:
- `VITE_API_BASE` - Optional API base URL for custom deployments

## Docker Compose

The Compose stack includes:
- `db` - Postgres 16
- `redis` - Redis 7
- `backend` - Go API server
- `frontend` - Static app served by nginx

## Project layout

- `backend/` - Go API service, data models, schema initialization
- `frontend/` - React/Vite SPA
- `docs/production.md` - production deployment guide

## Cleanup

- `make clean`

## Notes

- Update `DATABASE_DSN` and secrets before deploying to production.
- Review CORS settings and API exposure for public deployment.
- Use a secure secrets solution for `AISTUDIO_KEY` and database credentials.
