package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/models"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"github.com/redis/go-redis/v9"
)

// CheckHandler evaluates rate limits with config resolution.
func CheckHandler(s store.Store, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req models.CheckRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		if req.ClientID == "" {
			writeErr(w, http.StatusBadRequest, "client_id is required")
			return
		}

		algo, limit, windowSecs, ruleID, err := resolveConfig(
			r.Context(),
			client,
			req,
		)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}

		limiter, err := algorithms.NewLimiter(
			algo,
			s,
			limit,
			windowSecs,
		)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}

		allowed, remaining, err := limiter.Allow(
			r.Context(),
			req.ClientID,
		)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(models.CheckResponse{
			Allowed:   allowed,
			Remaining: remaining,
			Algorithm: algo,
			ClientID:  req.ClientID,
			RuleID:    ruleID,
		})
	}

}

// resolveConfig returns algorithm/limit/windowSecs from stored rule or request body.
func resolveConfig(
	ctx context.Context,
	client *redis.Client,
	req models.CheckRequest,
) (algo string, limit, windowSecs int, ruleID string, err error) {

	// Try to find a stored rule for this client
	clientKey := fmt.Sprintf("rule:by-client:%s", req.ClientID)

	storedRuleID, redisErr := client.Get(ctx, clientKey).Result()
	if redisErr == nil && storedRuleID != "" {

		// Stored rule found
		ruleKey := fmt.Sprintf("rule:%s", storedRuleID)

		fields, hErr := client.HGetAll(ctx, ruleKey).Result()
		if hErr == nil && len(fields) > 0 {

			l, _ := strconv.Atoi(fields["limit"])
			ws, _ := strconv.Atoi(fields["window_secs"])

			return fields["algorithm"], l, ws, storedRuleID, nil
		}
	}

	// No stored rule - require fields in request body
	if req.Algorithm == "" {
		return "", 0, 0, "", errors.New("algorithm is required")
	}

	if req.Limit <= 0 {
		return "", 0, 0, "", errors.New("limit must be > 0")
	}

	if req.WindowSecs <= 0 {
		return "", 0, 0, "", errors.New("window_secs must be > 0")
	}

	return req.Algorithm, req.Limit, req.WindowSecs, "", nil
}

// writeErr writes a JSON error response.
func writeErr(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error: msg,
	})
}
