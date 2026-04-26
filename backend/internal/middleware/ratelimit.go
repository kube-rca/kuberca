// Rate-limit middleware using a per-remote-IP token bucket.
//
// Backed by golang.org/x/time/rate. The middleware keeps an in-memory map of
// limiters keyed by client IP. Stale entries are evicted lazily once they
// exceed an inactivity threshold to bound memory usage. This is sufficient
// for a single-replica deployment (current SSE constraint already enforces
// replica=1). For multi-replica setups operators should front the webhook
// with a shared rate limiter (e.g. ingress, API gateway).
package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitConfig configures RateLimitMiddleware.
//
// RequestsPerMinute is the steady-state allowance per client IP. Burst is the
// short-term allowance and defaults to RequestsPerMinute when zero.
// CleanupInterval and IdleTTL control eviction of inactive limiters.
type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
	CleanupInterval   time.Duration
	IdleTTL           time.Duration
}

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	limit    rate.Limit
	burst    int
	idleTTL  time.Duration
}

func newRateLimiterStore(perMinute, burst int, idleTTL time.Duration) *rateLimiterStore {
	if perMinute <= 0 {
		perMinute = 100
	}
	if burst <= 0 {
		burst = perMinute
	}
	if idleTTL <= 0 {
		idleTTL = 10 * time.Minute
	}
	return &rateLimiterStore{
		limiters: make(map[string]*ipLimiter),
		// rate.Limit is events per second.
		limit:   rate.Limit(float64(perMinute) / 60.0),
		burst:   burst,
		idleTTL: idleTTL,
	}
}

func (s *rateLimiterStore) get(ip string, now time.Time) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.limiters[ip]; ok {
		entry.lastSeen = now
		return entry.limiter
	}
	lim := rate.NewLimiter(s.limit, s.burst)
	s.limiters[ip] = &ipLimiter{limiter: lim, lastSeen: now}
	return lim
}

func (s *rateLimiterStore) cleanup(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ip, entry := range s.limiters {
		if now.Sub(entry.lastSeen) > s.idleTTL {
			delete(s.limiters, ip)
		}
	}
}

// RateLimitMiddleware returns a Gin middleware that enforces a per-IP token
// bucket. Requests over the limit receive HTTP 429.
func RateLimitMiddleware(cfg RateLimitConfig) gin.HandlerFunc {
	store := newRateLimiterStore(cfg.RequestsPerMinute, cfg.Burst, cfg.IdleTTL)

	cleanup := cfg.CleanupInterval
	if cleanup <= 0 {
		cleanup = 5 * time.Minute
	}
	go func() {
		ticker := time.NewTicker(cleanup)
		defer ticker.Stop()
		for now := range ticker.C {
			store.cleanup(now)
		}
	}()

	return func(c *gin.Context) {
		ip := clientIP(c)
		limiter := store.get(ip, time.Now())
		if !limiter.Allow() {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// clientIP returns the canonical remote IP. Gin's ClientIP already honours
// trusted proxies; we fall back to the raw RemoteAddr if it returns empty.
func clientIP(c *gin.Context) string {
	if ip := c.ClientIP(); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return host
}
