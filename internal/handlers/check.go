package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/models"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
)

// CheckHandler evaluates rate limits via HTTP.
func CheckHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Parse JSON
		var req models.CheckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.CheckResponse{
				Error: "Invalid JSON body",
			})
			return
		}

		// Validate required fields
		if req.ClientID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.CheckResponse{
				Error: "client_id is required",
			})
			return
		}

		// Create limiter using factory
		limiter, err := algorithms.NewLimiter(
			req.Algorithm,
			s,
			req.Limit,
			req.WindowSecs,
		)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.CheckResponse{
				Error: err.Error(),
			})
			return
		}

		// Execute Allow()
		allowed, remaining, err := limiter.Allow(
			r.Context(),
			req.ClientID,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.CheckResponse{
				Error: "Internal Server Error",
			})
			return
		}

		// Success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.CheckResponse{
			Allowed:   allowed,
			Remaining: remaining,
			Algorithm: req.Algorithm,
			ClientID:  req.ClientID,
		})
	}
}
