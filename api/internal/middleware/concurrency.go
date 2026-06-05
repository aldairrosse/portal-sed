package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// context key for the expected version extracted from If-Match.
type ctxKeyVersion struct{}

// ExpectedVersionFromContext returns the version extracted from the If-Match header.
// Returns 0 if not set.
func ExpectedVersionFromContext(ctx context.Context) int {
	v, _ := ctx.Value(ctxKeyVersion{}).(int)
	return v
}

// OptimisticLock extracts the If-Match header, parses the version integer,
// and injects it into the request context. Returns:
//   - 428 Precondition Required if If-Match is missing
//   - 400 Bad Request if the If-Match value is not a valid integer
func OptimisticLock(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ifMatch := r.Header.Get("If-Match")
		if ifMatch == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionRequired)
			ae := errors.NewAPIErrorResponse(errors.ErrMissingIfMatch, "")
			_, _ = w.Write(ae.MustMarshalJSON())
			return
		}

		// Strip surrounding quotes if present (e.g. "1")
		cleaned := strings.Trim(ifMatch, `"`)
		version, err := strconv.Atoi(cleaned)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			ae := errors.NewAPIErrorResponse(errors.ErrInvalidIfMatch, "")
			_, _ = w.Write(ae.MustMarshalJSON())
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyVersion{}, version)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
