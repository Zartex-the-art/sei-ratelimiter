package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/handlers"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"github.com/redis/go-redis/v9"
)

func testRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func TestCheckHandler_ValidRequest(t *testing.T) {
	fakeStore := store.NewFakeStore()
	handler := handlers.CheckHandler(fakeStore, testRedisClient())

	body := []byte(`{"client_id":"user-1","algorithm":"fixed_window","limit":5,"window_secs":60}`)

	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(body))
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
	fakeStore := store.NewFakeStore()
	handler := handlers.CheckHandler(fakeStore, testRedisClient())

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
	fakeStore := store.NewFakeStore()
	handler := handlers.CheckHandler(fakeStore, testRedisClient())

	body := []byte(`{"algorithm":"fixed_window","limit":5,"window_secs":60}`)

	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCheckHandler_InvalidAlgorithm(t *testing.T) {
	fakeStore := store.NewFakeStore()
	handler := handlers.CheckHandler(fakeStore, testRedisClient())

	body := []byte(`{"client_id":"u1","algorithm":"magic","limit":5,"window_secs":60}`)

	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCheckHandler_NegativeLimit(t *testing.T) {
	fakeStore := store.NewFakeStore()
	handler := handlers.CheckHandler(fakeStore, testRedisClient())

	body := []byte(`{"client_id":"u1","algorithm":"fixed_window","limit":-5,"window_secs":60}`)

	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCheckHandler_ZeroWindow(t *testing.T) {
	fakeStore := store.NewFakeStore()
	handler := handlers.CheckHandler(fakeStore, testRedisClient())

	body := []byte(`{"client_id":"u1","algorithm":"fixed_window","limit":5,"window_secs":0}`)

	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}