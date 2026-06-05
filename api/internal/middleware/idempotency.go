package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// IdempotencyStore is the interface for idempotency-key storage.
// Implementations can be in-memory (testing/dev) or Redis-backed (production).
type IdempotencyStore interface {
	// Get retrieves a cached idempotency entry. Returns nil, nil if not found.
	Get(ctx context.Context, key string) (*IdempotencyEntry, error)
	// Set stores an idempotency entry with the given TTL.
	Set(ctx context.Context, key string, entry *IdempotencyEntry, ttl time.Duration) error
}

// IdempotencyEntry holds the cached response for an idempotency key.
type IdempotencyEntry struct {
	StatusCode   int    `json:"status_code"`
	Body         []byte `json:"body"`
	PayloadHash  string `json:"payload_hash"` // SHA256 of the request body
}

// InMemoryIdempotencyStore is a simple in-memory store for testing/dev.
type InMemoryIdempotencyStore struct {
	mu   sync.Mutex
	data map[string]*cacheItem
}

type cacheItem struct {
	entry    *IdempotencyEntry
	expireAt time.Time
}

// NewInMemoryIdempotencyStore creates a new in-memory idempotency store.
func NewInMemoryIdempotencyStore() *InMemoryIdempotencyStore {
	return &InMemoryIdempotencyStore{
		data: make(map[string]*cacheItem),
	}
}

func (s *InMemoryIdempotencyStore) Get(_ context.Context, key string) (*IdempotencyEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[key]
	if !ok || time.Now().After(item.expireAt) {
		if ok {
			delete(s.data, key)
		}
		return nil, nil
	}
	return item.entry, nil
}

func (s *InMemoryIdempotencyStore) Set(_ context.Context, key string, entry *IdempotencyEntry, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &cacheItem{
		entry:    entry,
		expireAt: time.Now().Add(ttl),
	}
	return nil
}

// context key for the idempotency key value injected by middleware.
type ctxKeyIdempotency struct{}

// IdempotencyKeyFromContext returns the Idempotency-Key extracted from the request,
// or empty string if not set.
func IdempotencyKeyFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyIdempotency{}).(string)
	return v
}

// hashBody computes the SHA256 hex digest of the request body.
func hashBody(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

// Idempotency returns middleware that handles Idempotency-Key headers.
//
// Flow:
//  1. Extract Idempotency-Key from header; if missing, pass through.
//  2. If key exists in store:
//     - Compare payload hash; if match → return cached response.
//     - If mismatch → return 409 IDEMPOTENCY_KEY_CONFLICT.
//  3. If key not in store:
//     - Execute handler, buffer response.
//     - On 2xx, cache entry with TTL (default 24h).
//     - On 4xx/5xx, do not cache.
func Idempotency(store IdempotencyStore, ttl time.Duration) func(http.Handler) http.Handler {
	if store == nil {
		store = NewInMemoryIdempotencyStore()
	}
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Validate key is a UUID
			if _, err := uuid.Parse(key); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				de := errors.NewDomainError(errors.InvalidRequest,
					"Idempotency-Key must be a valid UUID v4", err)
				ae := errors.NewAPIErrorResponse(de, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			// Read request body for hashing
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body.Close()
			// Replace body for downstream handlers
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			payloadHash := hashBody(bodyBytes)
			redisKey := "idempotency:" + key

			// Check store
			existing, err := store.Get(r.Context(), redisKey)
			if err == nil && existing != nil {
				if existing.PayloadHash == payloadHash {
					// Same payload → return cached result
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(existing.StatusCode)
					_, _ = w.Write(existing.Body)
					return
				}
				// Different payload → conflict
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				ae := errors.NewAPIErrorResponse(errors.ErrIdempotencyConflict, "")
				_, _ = w.Write(ae.MustMarshalJSON())
				return
			}

			// Inject key into context
			ctx := context.WithValue(r.Context(), ctxKeyIdempotency{}, key)
			r = r.WithContext(ctx)

			// Wrap ResponseWriter to buffer the response
			lrw := &idempotencyResponseWriter{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(lrw, r)

			// Only cache 2xx responses
			if lrw.statusCode >= 200 && lrw.statusCode < 300 {
				entry := &IdempotencyEntry{
					StatusCode:  lrw.statusCode,
					Body:        lrw.body.Bytes(),
					PayloadHash: payloadHash,
				}
				_ = store.Set(r.Context(), redisKey, entry, ttl)
			}
		})
	}
}

// idempotencyResponseWriter buffers the response for caching.
type idempotencyResponseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *idempotencyResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *idempotencyResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Unwrap exposes the underlying http.ResponseWriter for compatibility.
func (w *idempotencyResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
