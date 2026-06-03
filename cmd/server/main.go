package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/config"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

type CheckRequest struct {
	ClientID  string `json:"client_id"`
	Algorithm string `json:"algorithm"`
	Limit     int    `json:"limit"`
	WindowSec int    `json:"window_secs"`
}

type CheckResponse struct {
	Allowed   bool   `json:"allowed"`
	Remaining int    `json:"remaining"`
	Algorithm string `json:"algorithm"`
}

func main() {
	cfg := config.Load()

	rs := store.NewRedisStore(cfg.RedisURL)

	if err := rs.Ping(context.Background()); err != nil {
		log.Printf("warning: Redis not reachable at %s: %v", cfg.RedisURL, err)
	} else {
		log.Printf("Redis connected at %s", cfg.RedisURL)
	}

	// Sanity check: all algorithms register correctly at startup
	for _, algo := range algorithms.ValidAlgorithms() {
		if _, err := algorithms.NewLimiter(algo, rs, 100, 60); err != nil {
			log.Fatalf("factory sanity check failed for %s: %v", algo, err)
		}
	}

	log.Printf(
		"Algorithm factory: all %d algorithms registered",
		len(algorithms.ValidAlgorithms()),
	)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","node":%q}`, cfg.NodeID)
	})

	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req CheckRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		limiter, err := algorithms.NewLimiter(
			req.Algorithm,
			rs,
			req.Limit,
			req.WindowSec,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		allowed, remaining, err := limiter.Allow(
			context.Background(),
			req.ClientID,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(CheckResponse{
			Allowed:   allowed,
			Remaining: remaining,
			Algorithm: req.Algorithm,
		})
	})

	log.Printf(
		"starting sei-ratelimiter node=%s port=%s",
		cfg.NodeID,
		cfg.Port,
	)

	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}

}
