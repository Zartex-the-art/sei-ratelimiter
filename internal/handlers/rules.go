package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// CreateRuleHandler handles POST /rules.
func CreateRuleHandler(client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req models.RuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid JSON body",
			})
			return
		}

		if req.ClientID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "client_id is required",
			})
			return
		}

		if req.Limit <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "limit must be > 0",
			})
			return
		}

		if req.WindowSecs <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "window_secs must be > 0",
			})
			return
		}

		valid := false
		for _, a := range algorithms.ValidAlgorithms() {
			if req.Algorithm == a {
				valid = true
				break
			}
		}

		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("unknown algorithm %q", req.Algorithm),
			})
			return
		}

		rule := models.Rule{
			ID:         uuid.New().String(),
			ClientID:   req.ClientID,
			Algorithm:  req.Algorithm,
			Limit:      req.Limit,
			WindowSecs: req.WindowSecs,
			Enabled:    req.Enabled,
			CreatedAt:  time.Now().UTC(),
		}

		ctx := r.Context()

		ruleKey := fmt.Sprintf("rule:%s", rule.ID)
		clientKey := fmt.Sprintf("rule:by-client:%s", rule.ClientID)
		indexKey := "rules:index"

		pipe := client.Pipeline()

		pipe.HSet(ctx, ruleKey, map[string]interface{}{
			"id":          rule.ID,
			"client_id":   rule.ClientID,
			"algorithm":   rule.Algorithm,
			"limit":       strconv.Itoa(rule.Limit),
			"window_secs": strconv.Itoa(rule.WindowSecs),
			"enabled":     strconv.FormatBool(rule.Enabled),
			"created_at":  rule.CreatedAt.Format(time.RFC3339),
		})

		pipe.SAdd(ctx, indexKey, rule.ID)
		pipe.Set(ctx, clientKey, rule.ID, 0)

		if _, err := pipe.Exec(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "failed to store rule",
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rule)
	}
}

// ListRulesHandler handles GET /rules.
func ListRulesHandler(client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx := r.Context()

		ids, err := client.SMembers(ctx, "rules:index").Result()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "failed to fetch rules",
			})
			return
		}

		rules := make([]models.Rule, 0)

		for _, id := range ids {
			fields, err := client.HGetAll(ctx, fmt.Sprintf("rule:%s", id)).Result()
			if err != nil || len(fields) == 0 {
				continue
			}

			rule, err := ruleFromHash(fields)
			if err != nil {
				continue
			}

			rules = append(rules, rule)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.RulesListResponse{
			Rules: rules,
		})
	}
}

// GetRuleHandler handles GET /rules/{id}
func GetRuleHandler(client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.PathValue("id")
		key := fmt.Sprintf("rule:%s", id)

		fields, err := client.HGetAll(r.Context(), key).Result()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "failed to fetch rule",
			})
			return
		}

		if len(fields) == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rule not found",
			})
			return
		}

		rule, err := ruleFromHash(fields)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "failed to parse rule",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rule)
	}
}

// DeleteRuleHandler handles DELETE /rules/{id}
func DeleteRuleHandler(client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GetRuleHandler HIT")
		w.Header().Set("Content-Type", "application/json")

		id := r.PathValue("id")
		ruleKey := fmt.Sprintf("rule:%s", id)

		fields, err := client.HGetAll(r.Context(), ruleKey).Result()
		if err != nil || len(fields) == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rule not found",
			})
			return
		}

		clientID := fields["client_id"]
		clientKey := fmt.Sprintf("rule:by-client:%s", clientID)

		pipe := client.Pipeline()
		pipe.Del(r.Context(), ruleKey)
		pipe.SRem(r.Context(), "rules:index", id)
		pipe.Del(r.Context(), clientKey)

		if _, err := pipe.Exec(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "failed to delete rule",
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ruleFromHash converts a Redis hash to a Rule struct.
func ruleFromHash(fields map[string]string) (models.Rule, error) {
	limit, _ := strconv.Atoi(fields["limit"])
	windowSecs, _ := strconv.Atoi(fields["window_secs"])
	enabled, _ := strconv.ParseBool(fields["enabled"])
	createdAt, _ := time.Parse(time.RFC3339, fields["created_at"])

	return models.Rule{
		ID:         fields["id"],
		ClientID:   fields["client_id"],
		Algorithm:  fields["algorithm"],
		Limit:      limit,
		WindowSecs: windowSecs,
		Enabled:    enabled,
		CreatedAt:  createdAt,
	}, nil
}
