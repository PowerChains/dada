package main

import (
	context "context"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := LoadConfig()
	ctx := context.Background()

	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to open postgres: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("postgres unreachable: %v", err)
	}

	cache := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := cache.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis unreachable: %v", err)
	}

	if err := initSchema(ctx, db); err != nil {
		log.Fatalf("failed init schema: %v", err)
	}

	server := NewServer(cfg, db, cache)
	go server.RunArbitrageBroadcaster(ctx)

	addr := ":" + cfg.Port
	log.Printf("DADA backend listening on %s", addr)
	if err := http.ListenAndServe(addr, server.SetupRouter()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
