package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCheckRequestJSON(t *testing.T) {
	_, err := json.Marshal(CheckRequest{
		ClientID:   "c1",
		Algorithm:  "fixed_window",
		Limit:      5,
		WindowSecs: 60,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckResponseJSON(t *testing.T) {
	_, err := json.Marshal(CheckResponse{
		Allowed:   true,
		Remaining: 4,
		ClientID:  "c1",
		Algorithm: "fixed_window",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestErrorResponseJSON(t *testing.T) {
	_, err := json.Marshal(ErrorResponse{
		Error: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRuleJSON(t *testing.T) {
	_, err := json.Marshal(Rule{
		ID:         "1",
		ClientID:   "c1",
		Algorithm:  "fixed_window",
		Limit:      5,
		WindowSecs: 60,
		Enabled:    true,
		CreatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRulesListResponseJSON(t *testing.T) {
	_, err := json.Marshal(RulesListResponse{
		Rules: []Rule{
			{
				ID:       "1",
				ClientID: "c1",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
