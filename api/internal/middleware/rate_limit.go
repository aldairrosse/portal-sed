package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// RateLimitStore is the interface for rate-limit counters.
// Implementations can be in-memory (default) or Redis-backed.
type RateLimitStore interface {
	// Increment returns the new count after incrementing the counter for the given key.
	// If the window has passed, the counter resets.
	Increment(ctx context.Context, key string, window time.Duration, max int) (int, error)
}

// InMemoryRateLimitStore is a simple in-memory token-bucket implementation.
// Not suitable for production multi-replica deployments; use Redis instead.
type InMemoryRateLimitStore struct {
	mu       sync.Mutex
	counters map[string]*bucketEntry
}

type bucketEntry struct {
	count       int
	windowStart time.Time
}

// NewInMemoryRateLimitStore creates a new in-memory rate-limit store.
func NewInMemoryRateLimitStore() *InMemoryRateLimitStore {
	return &InMemoryRateLimitStore{
		counters: make(map[string]*bucketEntry),
	}
}

func (s *InMemoryRateLimitStore) Increment(_ context.Context, key string, window time.Duration, max int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry, exists := s.counters[key]

	if !exists || now.Sub(entry.windowStart) > window {
		// new window
		s.counters[key] = &bucketEntry{count: 1, windowStart: now}
		return 1, nil
	}

	entry.count++
	return entry.count, nil
}

// RateLimitConfig holds the per-operation rate-limit configuration.
type RateLimitConfig struct {
	Window   time.Duration
	MaxCount int
	Store    RateLimitStore
}

// RateLimit returns an HTTP middleware that enforces a token-bucket rate limit
// per organization. The key is derived from the org ID and an operation type
// (read|write). It sets X-RateLimit-* response headers.
//
// Default store is in-memory if cfg.Store is nil.
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.Store == nil {
		cfg.Store = NewInMemoryRateLimitStore()
	}
	if cfg.Window == 0 {
		cfg.Window = time.Minute
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := OrgIDFromContext(r.Context())
			if orgID == "" {
				orgID = "unknown"
			}

			// Determine operation type from method
			opType := "write"
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				opType = "read"
			}

			key := "ratelimit:" + opType + ":" + orgID
			count, err := cfg.Store.Increment(r.Context(), key, cfg.Window, cfg.MaxCount)
			if err != nil {
				// If store fails, allow the request but log
				next.ServeHTTP(w, r)
				return
			}

			// Calculate remaining and reset
			remaining := cfg.MaxCount - count
			if remaining < 0 {
				remaining = 0
			}
			resetSecs := int(cfg.Window.Seconds())

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.MaxCount))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(resetSecs))

			if count > cfg.MaxCount {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				de := errors.ErrRateLimitExceeded.WithDetails(
					"limit: "+strconv.Itoa(cfg.MaxCount),
					"window: "+cfg.Window.String(),
					"key: "+key,
				)
				ae := errors.NewAPIErrorResponse(de, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
