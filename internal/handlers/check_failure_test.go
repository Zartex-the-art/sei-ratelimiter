package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/handlers"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/models"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/testhelpers"
)

// TestCheckHandler_RedisDown verifies that a Redis failure returns 503
// and NOT 500 or a hang.
func TestCheckHandler_RedisDown(t *testing.T) {
	// Use ErrorStore to simulate Redis being unreachable
	errorStore := store.NewErrorStore()
	// Use a real Redis client for rule lookup (but no stored rules)
	client := testhelpers.RedisClient(t)
	handler := handlers.CheckHandler(errorStore, client)
	body := []byte(`{"client_id":"redis-down","algorithm":"fixed_window","limit": 5,"window_secs":60}`)
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	// Must return 503 Service Unavailable — not 500
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Redis down: expected 503, got %d — body: %s", w.Code, w.Body.String())
	}
	// Body must have error field
	var resp models.CheckResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error == "" {
		t.Error("503 response must include error field")
	}
	t.Logf("Redis down response: %s", w.Body.String())
}

// TestCheckHandler_RedisDown_AllAlgorithms verifies 503 for all three algorithms.
func TestCheckHandler_RedisDown_AllAlgorithms(t *testing.T) {
	client := testhelpers.RedisClient(t)
	errorStore := store.NewErrorStore()
	handler := handlers.CheckHandler(errorStore, client)
	algorithms := []string{"fixed_window", "sliding_window", "token_bucket"}
	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			bodyStr := `{"client_id":"down-test","algorithm":"` + algo + `","limit": 5,"window_secs":60}`
			req := httptest.NewRequest(http.MethodPost, "/check",
				bytes.NewBufferString(bodyStr))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusServiceUnavailable {
				t.Errorf("%s: expected 503, got %d", algo, w.Code)
			}
		})
	}
}

// TestCheckHandler_NoPanicOnRedisDown verifies the handler never panics
// when Redis is unavailable.
func TestCheckHandler_NoPanicOnRedisDown(t *testing.T) {
	client := testhelpers.RedisClient(t)
	errorStore := store.NewErrorStore()
	handler := handlers.CheckHandler(errorStore, client)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handler panicked on Redis error: %v", r)
		}
	}()
	// Send many requests — none should panic
	for i := 0; i < 20; i++ {
		body := []byte(`{"client_id":"no-panic","algorithm":"fixed_window","limit": 5,"window_secs":60}`)
		req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		// Just verify it doesn't panic — status code check is secondary
		if w.Code != 503 && w.Code != 200 {
			t.Logf("request %d: unexpected status %d", i+1, w.Code)
		}
	}
}

// TestCheckHandler_RecoveryAfterRedisError verifies the handler works correctly
// when a working store replaces the error store (simulates Redis recovery).
func TestCheckHandler_RecoveryAfterRedisError(t *testing.T) {
	client := testhelpers.RedisClient(t)
	// First: use error store — get 503
	handler := handlers.CheckHandler(store.NewErrorStore(), client)
	body := []byte(`{"client_id":"recovery","algorithm":"fixed_window","limit": 5,"window_secs":60}`)
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 503 {
		t.Errorf("pre-recovery: expected 503, got %d", w.Code)
	}
	// Then: use working store — get 200
	handler2 := handlers.CheckHandler(store.NewFakeStore(), client)
	req2 := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(body))
	w2 := httptest.NewRecorder()
	handler2.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Errorf("post-recovery: expected 200, got %d", w2.Code)
	}
	var resp models.CheckResponse
	json.Unmarshal(w2.Body.Bytes(), &resp)
	if !resp.Allowed {
		t.Error("post-recovery: should be allowed")
	}
}
