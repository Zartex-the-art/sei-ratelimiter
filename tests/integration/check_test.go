package integration

import (
	"net/http"
	"testing"
)

func apiBase(t *testing.T) string {
	base := "http://localhost:8080"
	resp, err := http.Get(base + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skipf("API not reachable - skipping integration test")
	}
	return base
}

func TestCheckEndpoint_Scaffold(t *testing.T) {
	t.Skip("POST /check not implemented yet - scaffolded for Phase 3")
}
