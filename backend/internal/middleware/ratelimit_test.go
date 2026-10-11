package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	if !rl.Allow("client") {
		t.Fatal("First request should be allowed")
	}
	if !rl.Allow("client") {
		t.Fatal("Second request should be allowed")
	}
	if rl.Allow("client") {
		t.Fatal("Third request should be blocked")
	}
	if !rl.Allow("other-client") {
		t.Fatal("A different client should not be affected")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(1, 40*time.Millisecond)

	if !rl.Allow("client") {
		t.Fatal("First request should be allowed")
	}
	if rl.Allow("client") {
		t.Fatal("Second request should be blocked within the window")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("client") {
		t.Fatal("Request should be allowed after the window expires")
	}
}

func TestRateLimitMiddleware_BlocksWith429(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("Expected 200 on first request, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected 429 on second request, got %d", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Error("Expected Retry-After header on 429")
	}
}

func TestGetClientID_StripsPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.16.158.222:54321"

	if got := getClientID(req); got != "172.16.158.222" {
		t.Errorf("Expected host without port, got %q", got)
	}
}
