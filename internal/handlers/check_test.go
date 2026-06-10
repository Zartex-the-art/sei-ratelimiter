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
	"github.com/redis/go-redis/v9"
)

func newHandler() http.HandlerFunc {
	fakeStore := store.NewFakeStore()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return handlers.CheckHandler(fakeStore, client)

}

func TestCheckHandler_ValidRequest(t *testing.T) {
	handler := newHandler()

	body := []byte(`{"client_id":"user-1","algorithm":"fixed_window","limit":5,"window_secs":60}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/check",
		bytes.NewBuffer(body),
	)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["allowed"] != true {
		t.Errorf("expected allowed=true")
	}

}

func TestCheckHandler_InvalidJSON(t *testing.T) {
	handler := newHandler()

	req := httptest.NewRequest(
		http.MethodPost,
		"/check",
		bytes.NewBuffer([]byte(`{bad}`)),
	)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

}

func TestCheckHandler_MissingClientID(t *testing.T) {
	handler := newHandler()

	body := []byte(`{"algorithm":"fixed_window","limit":5,"window_secs":60}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/check",
		bytes.NewBuffer(body),
	)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

}

func TestCheckHandler_InvalidAlgorithm(t *testing.T) {
	handler := newHandler()

	body := []byte(`{"client_id":"u1","algorithm":"magic","limit":5,"window_secs":60}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/check",
		bytes.NewBuffer(body),
	)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

}

func TestCheckHandler_NegativeLimit(t *testing.T) {
	handler := newHandler()

	body := []byte(`{"client_id":"u1","algorithm":"fixed_window","limit":-5,"window_secs":60}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/check",
		bytes.NewBuffer(body),
	)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

}

func TestCheckHandler_ZeroWindow(t *testing.T) {
	handler := newHandler()

	body := []byte(`{"client_id":"u1","algorithm":"fixed_window","limit":5,"window_secs":0}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/check",
		bytes.NewBuffer(body),
	)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

}
func TestCheckHandler_VeryLargeLimit(t *testing.T) {
	client := testhelpers.RedisClient(t)

	defer testhelpers.FlushKeys(
		t,
		client,
		"fw:large-limit",
	)

	handler := handlers.CheckHandler(
		store.NewFakeStore(), client,
	)

	body := []byte(
		`{"client_id":"large-limit","algorithm":"fixed_window","limit":1000000,"window_secs":3600}`,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/check",
		bytes.NewBuffer(body),
	)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.CheckResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.Allowed {
		t.Error("expected request to be allowed")
	}

	if resp.Remaining != 999999 {
		t.Errorf(
			"remaining: got %d, want %d",
			resp.Remaining,
			999999,
		)
	}
}
