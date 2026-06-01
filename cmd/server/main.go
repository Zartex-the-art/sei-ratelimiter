package main

import (

	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/config"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

func main() {
	cfg := config.Load()
	rs := store.NewRedisStore(cfg.RedisURL)
	if err := rs.Ping(context.Background()); err != nil {
		log.Printf("warning: Redis not reachable at %s: %v", cfg.RedisURL, err)
	} else {
		log.Printf("Redis connected at %s", cfg.RedisURL)
	}
	// FixedWindow wired with Redis store
	// Algorithms expand in Phase 3 with the factory pattern
	_ = algorithms.NewFixedWindow(100, 60, rs)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","node":%q}`, cfg.NodeID)
	})
	log.Printf("starting node=%s port=%s", cfg.NodeID, cfg.Port)
        if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
                log.Fatalf("server error: %v", err)
        }
        }
