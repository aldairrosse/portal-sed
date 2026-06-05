// Package cursor provides cursor-based pagination support using a base64-encoded
// (id, updated_at) tuple. This avoids the performance degradation of OFFSET-based
// pagination on large datasets.
package cursor

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// Cursor represents a pagination cursor as a (ID, UpdatedAt) composite key.
type Cursor struct {
	ID        uuid.UUID `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Encode serialises the cursor to a base64-encoded JSON string.
func (c *Cursor) Encode() (string, error) {
	raw, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(raw), nil
}

// MustEncode is like Encode but panics on error.
func (c *Cursor) MustEncode() string {
	s, err := c.Encode()
	if err != nil {
		panic("cursor: encode failed: " + err.Error())
	}
	return s
}

// DecodeCursor parses a base64-encoded cursor string into a Cursor.
// Returns an *errors.DomainError with code INVALID_REQUEST if the input
// is malformed.
func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}
	raw, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest,
			"cursor: invalid base64 encoding", err)
	}
	var c Cursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest,
			"cursor: invalid JSON in cursor payload", err)
	}
	return &c, nil
}

// PaginatedList is a generic wrapper for cursor-paginated responses.
type PaginatedList[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination holds the cursor-based pagination metadata.
type Pagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// NewPaginatedList creates a PaginatedList from query results.
// It expects `results` to be at most limit+1 items. If len(results) > limit,
// the extra item is removed, has_more is set, and the next cursor is encoded
// from the last remaining item.
//
// getCursor is a callback that extracts the (id, updated_at) tuple from an element.
func NewPaginatedList[T any](
	results []T,
	limit int,
	getCursor func(T) (uuid.UUID, time.Time),
) (*PaginatedList[T], error) {
	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit]
	}

	pl := &PaginatedList[T]{
		Data:       results,
		Pagination: Pagination{HasMore: hasMore},
	}

	if hasMore && len(results) > 0 {
		last := results[len(results)-1]
		id, updatedAt := getCursor(last)
		c := &Cursor{ID: id, UpdatedAt: updatedAt}
		next, err := c.Encode()
		if err != nil {
			return nil, err
		}
		pl.Pagination.NextCursor = next
	}

	return pl, nil
}
