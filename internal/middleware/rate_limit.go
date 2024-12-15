package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	requests map[string]int
	mu       sync.Mutex
}

var limiter = &rateLimiter{
	requests: make(map[string]int),
}

func RateLimit(maxRequests int, duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)

			limiter.mu.Lock()
			defer limiter.mu.Unlock()

			// Reset counters periodically
			go func() {
				time.Sleep(duration)
				limiter.mu.Lock()
				delete(limiter.requests, ip)
				limiter.mu.Unlock()
			}()

			// Increment request count
			limiter.requests[ip]++
			if limiter.requests[ip] > maxRequests {
				http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
