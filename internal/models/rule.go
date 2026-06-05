package models

import "time"

// RuleRequest is the JSON body for POST /rules.
type RuleRequest struct {
	ClientID   string `json:"client_id"`
	Algorithm  string `json:"algorithm"`
	Limit      int    `json:"limit"`
	WindowSecs int    `json:"window_secs"`
	Enabled    bool   `json:"enabled"`
}

// Rule is the full rule object stored in Redis and returned by the API.
type Rule struct {
	ID         string    `json:"id"`
	ClientID   string    `json:"client_id"`
	Algorithm  string    `json:"algorithm"`
	Limit      int       `json:"limit"`
	WindowSecs int       `json:"window_secs"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
}

// RulesListResponse wraps the rules array.
type RulesListResponse struct {
	Rules []Rule `json:"rules"`
}
