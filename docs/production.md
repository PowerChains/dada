# Production Deployment Guide

This document describes the production-ready deployment setup for the DADA app.

## Architecture

- `backend/` contains the Go API server and database access layer.
- `frontend/` contains the React/Vite UI.
- `docker-compose.yml` orchestrates Postgres, Redis, backend, and frontend.
- `frontend/nginx.conf` proxies API and websocket traffic to the backend.

## Production build

The production setup uses containerized services:

- `db`: Postgres 16
- `redis`: Redis 7
- `backend`: built Go binary with optional AI support
- `frontend`: static SPA served via nginx

## Build and run

From the repository root:

```bash
make setup
make docker-build
make docker-up
```

Or directly with Docker Compose:

```bash
docker compose up --build
```

The frontend will be exposed on `http://localhost:3000` and the backend API on `http://localhost:8080`.

## Environment variables

### Backend

- `DATABASE_DSN` - Postgres DSN, default: `postgres://dada:password@db:5432/dada?sslmode=disable`
- `REDIS_ADDR` - Redis address, default: `redis:6379`
- `PORT` - Backend listen port, default: `8080`
- `AISTUDIO_KEY` - Optional AI assistant API key

### Frontend

- `VITE_API_BASE` - Optional API base URL. In production, the frontend uses relative paths and nginx proxying.

## Production notes

- The `frontend` service uses nginx to route `/api/` and `/ws/` to the backend so the SPA can run from a single origin.
- The backend uses `CGO_ENABLED=0` for a portable static build.

## Database initialization

The backend initializes the database schema automatically on startup using `initSchema`.

## Security and configuration

- Do not use `postgres://dada:password@...` in production. Replace credentials with strong secrets.
- Use a secret manager or environment configuration for `AISTUDIO_KEY`.
- Lock down CORS origins before deploying publicly.
- Monitor and rotate keys and connection strings.
