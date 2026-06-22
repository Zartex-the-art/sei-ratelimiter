package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
"strings"
	"testing"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/handlers"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/testhelpers"
	"github.com/redis/go-redis/v9"
)

func redisClientForTests(t *testing.T) *redis.Client {
	t.Helper()

	client := testhelpers.RedisClient(t)
	return client
}

func TestRulesHandler_CreateReturns201(t *testing.T) {
	client := redisClientForTests(t)
	defer testhelpers.FlushKeys(t, client, "rule:*", "rules:index")

	handler := handlers.CreateRuleHandler(client)

	body := []byte(`{"client_id":"alice","algorithm":"fixed_window","limit":5,"window_secs":60,"enabled":true}`)

	req := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestRulesHandler_CreateReturnsFullObject(t *testing.T) {
	client := redisClientForTests(t)
	defer testhelpers.FlushKeys(t, client, "rule:*", "rules:index", "rule:by-client:bob")

	handler := handlers.CreateRuleHandler(client)

	body := []byte(`{"client_id":"bob","algorithm":"sliding_window","limit":20,"window_secs":30,"enabled":true}`)

	req := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var rule map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &rule)

	required := []string{
		"id",
		"client_id",
		"algorithm",
		"limit",
		"window_secs",
		"enabled",
		"created_at",
	}

	for _, field := range required {
		if _, ok := rule[field]; !ok {
			t.Errorf("response missing field: %s", field)
		}
	}

	if rule["client_id"] != "bob" {
		t.Errorf("client_id mismatch")
	}

	if rule["algorithm"] != "sliding_window" {
		t.Errorf("algorithm mismatch")
	}

	if rule["limit"].(float64) != 20 {
		t.Errorf("limit mismatch")
	}

	if rule["id"] == "" {
		t.Error("id must not be empty")
	}
}

func TestRulesHandler_ListReturnsEmpty(t *testing.T) {
	client := redisClientForTests(t)

	handler := handlers.ListRulesHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	rules, ok := resp["rules"]
	if !ok {
		t.Fatal("response missing 'rules' field")
	}

	rulesSlice, ok := rules.([]interface{})
	if !ok {
		t.Fatal("'rules' field must be an array")
	}

	if len(rulesSlice) != 0 {
		t.Errorf("expected empty array, got %d rules", len(rulesSlice))
	}
}

func TestRulesHandler_ListReturnsAll(t *testing.T) {
	client := redisClientForTests(t)
	defer testhelpers.FlushKeys(t, client, "rule:*", "rules:index", "rule:by-client:*")

	createHandler := handlers.CreateRuleHandler(client)

	clients := []string{
		"user-1",
		"user-2",
		"user-3",
	}

	for _, c := range clients {
		body := []byte(`{"client_id":"` + c + `","algorithm":"fixed_window","limit":5,"window_secs":60,"enabled":true}`)

		req := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		createHandler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("create failed for %s: %d", c, w.Code)
		}
	}

	listHandler := handlers.ListRulesHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)
	w := httptest.NewRecorder()

	listHandler.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	rules := resp["rules"].([]interface{})

	if len(rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(rules))
	}
}

func TestRulesHandler_ValidationErrors(t *testing.T) {
	client := redisClientForTests(t)

	handler := handlers.CreateRuleHandler(client)

	tests := []struct {
		name string
		body string
	}{
		{
			"missing client_id",
			`{"algorithm":"fixed_window","limit":5,"window_secs":60}`,
		},
		{
			"unknown algorithm",
			`{"client_id":"u1","algorithm":"unknown","limit":5,"window_secs":60}`,
		},
		{
			"limit zero",
			`{"client_id":"u1","algorithm":"fixed_window","limit":0,"window_secs":60}`,
		},
		{
			"limit negative",
			`{"client_id":"u1","algorithm":"fixed_window","limit":-1,"window_secs":60}`,
		},
		{
			"window_secs zero",
			`{"client_id":"u1","algorithm":"fixed_window","limit":5,"window_secs":0}`,
		},
		{
			"invalid json",
			`{bad json}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/rules",
				bytes.NewBuffer([]byte(tt.body)),
			)

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("%s: expected 400, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestRulesHandler_CreateDuplicateReturns409(t *testing.T) {
	client := redisClientForTests(t)

	handler := handlers.CreateRuleHandler(client)

	body := `{
		"client_id":"duplicate-client",
		"algorithm":"fixed_window",
		"limit":10,
		"window_secs":60,
		"enabled":true
	}`

	req1 := httptest.NewRequest(http.MethodPost, "/rules",
		strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/rules",
		strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusConflict {
		t.Fatalf("expected 409 got %d", rr2.Code)
	}
}

func TestRulesHandler_GetReturnsRule(t *testing.T) {
	client := redisClientForTests(t)
	defer testhelpers.FlushKeys(t, client, "rule:*", "rules:index", "rule:by-client:*")

	createHandler := handlers.CreateRuleHandler(client)

	body := []byte(`{"client_id":"get-user","algorithm":"fixed_window","limit":5,"window_secs":60,"enabled":true}`)

	createReq := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(body))
	createW := httptest.NewRecorder()

	createHandler.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)

	id := created["id"].(string)

	getHandler := handlers.GetRuleHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/rules/"+id, nil)
	req.SetPathValue("id", id)

	w := httptest.NewRecorder()

	getHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var rule map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &rule)

	if rule["id"] != id {
		t.Errorf("expected id %s, got %v", id, rule["id"])
	}

}

func TestRulesHandler_GetReturns404WhenMissing(t *testing.T) {
	client := redisClientForTests(t)

	handler := handlers.GetRuleHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/rules/missing", nil)
	req.SetPathValue("id", "missing")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

}

func TestRulesHandler_DeleteReturns204(t *testing.T) {
	client := redisClientForTests(t)
	defer testhelpers.FlushKeys(t, client, "rule:*", "rules:index", "rule:by-client:*")

	createHandler := handlers.CreateRuleHandler(client)

	body := []byte(`{"client_id":"delete-user","algorithm":"fixed_window","limit":5,"window_secs":60,"enabled":true}`)

	createReq := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(body))
	createW := httptest.NewRecorder()

	createHandler.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)

	id := created["id"].(string)

	deleteHandler := handlers.DeleteRuleHandler(client)

	req := httptest.NewRequest(http.MethodDelete, "/rules/"+id, nil)
	req.SetPathValue("id", id)

	w := httptest.NewRecorder()

	deleteHandler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

}

func TestRulesHandler_DeleteReturns404WhenMissing(t *testing.T) {
	client := redisClientForTests(t)

	handler := handlers.DeleteRuleHandler(client)

	req := httptest.NewRequest(http.MethodDelete, "/rules/missing", nil)
	req.SetPathValue("id", "missing")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

}
func TestRulesHandler_DeleteRemovesRule(t *testing.T) {
	client := redisClientForTests(t)
	defer testhelpers.FlushKeys(t, client, "rule:*", "rules:index", "rule:by-client:*")

	createHandler := handlers.CreateRuleHandler(client)

	body := []byte(`{"client_id":"delete-check","algorithm":"fixed_window","limit":5,"window_secs":60,"enabled":true}`)

	createReq := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(body))
	createW := httptest.NewRecorder()

	createHandler.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)

	id := created["id"].(string)

	deleteHandler := handlers.DeleteRuleHandler(client)

	req := httptest.NewRequest(http.MethodDelete, "/rules/"+id, nil)
	req.SetPathValue("id", id)

	w := httptest.NewRecorder()

	deleteHandler.ServeHTTP(w, req)

	getHandler := handlers.GetRuleHandler(client)

	getReq := httptest.NewRequest(http.MethodGet, "/rules/"+id, nil)
	getReq.SetPathValue("id", id)

	getW := httptest.NewRecorder()

	getHandler.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", getW.Code)
	}
}
func TestRulesHandler_ListSkipsMissingRules(t *testing.T) {
	client := redisClientForTests(t)
	defer testhelpers.FlushKeys(t, client, "rule:*", "rules:index")

	// Add an ID to the index without creating the corresponding rule hash.
	client.SAdd(
		context.Background(),
		"rules:index",
		"missing-rule",
	)

	handler := handlers.ListRulesHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	rules := resp["rules"].([]interface{})

	if len(rules) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(rules))
	}

}

