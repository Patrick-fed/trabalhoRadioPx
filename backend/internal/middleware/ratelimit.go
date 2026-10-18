package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[clientID]
	validRequests := make([]time.Time, 0)

	for _, req := range requests {
		if req.After(windowStart) {
			validRequests = append(validRequests, req)
		}
	}

	if len(validRequests) >= rl.limit {
		return false
	}

	rl.requests[clientID] = append(validRequests, now)
	return true
}

func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for clientID, requests := range rl.requests {
		validRequests := make([]time.Time, 0)
		for _, req := range requests {
			if req.After(windowStart) {
				validRequests = append(validRequests, req)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, clientID)
		} else {
			rl.requests[clientID] = validRequests
		}
	}
}

func (rl *RateLimiter) StartCleanupRoutine() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			rl.Cleanup()
		}
	}()
}

func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := getClientID(r)

			if !limiter.Allow(clientID) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Muitas tentativas. Tente novamente mais tarde."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientID returns the IP of the actual connection. Forwarding headers
// (X-Forwarded-For/X-Real-IP) are intentionally ignored: the server sits
// directly on the network, so trusting client-supplied headers would let an
// attacker rotate them and bypass the rate limit entirely.
func getClientID(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
