package limiter

import (
	"testing"
	"time"
)

func TestFixedWindowLimiter(t *testing.T) {
	limiter := NewFixedWindowLimiter(3, 5*time.Second)

	if !limiter.Allow("hari") {
		t.Error("request 1 should pass")
	}

	if !limiter.Allow("hari") {
		t.Error("request 2 should pass")
	}

	if !limiter.Allow("hari") {
		t.Error("request 3 should pass")
	}

	if limiter.Allow("hari") {
		t.Error("request 4 should fail")
	}
}
