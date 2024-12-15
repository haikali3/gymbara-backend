package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	requests sync.Map
	mu       sync.Mutex
}

type requestData struct {
	count      int
	lastAccess time.Time
}

var limiter = &rateLimiter{}

// limits the number of requests from a single IP address
func RateLimit(maxRequests int, duration time.Duration) func(http.Handler) http.Handler {
	//clean old entries
	go cleanupExpiredEntries(duration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			now := time.Now()

			// get or initialize the request data for IP
			value, _ := limiter.requests.LoadOrStore(ip, &requestData{
				count:      0,
				lastAccess: now,
			})
			requestData := value.(*requestData)

			// sync access to requestData
			limiter.mu.Lock()
			defer limiter.mu.Unlock()

			// reset counters periodically
			if now.Sub(requestData.lastAccess) > duration {
				requestData.count = 0
				requestData.lastAccess = now
			}

			requestData.count++
			if requestData.count > maxRequests {
				http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}

			limiter.requests.Store(ip, requestData)
			next.ServeHTTP(w, r)
		})
	}
}

func cleanupExpiredEntries(duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		<-ticker.C
		now := time.Now()
		limiter.requests.Range(func(key, value interface{}) bool {
			requestData := value.(*requestData)
			if now.Sub(requestData.lastAccess) > duration {
				limiter.requests.Delete(key)
			}
			return true
		})
	}
}
