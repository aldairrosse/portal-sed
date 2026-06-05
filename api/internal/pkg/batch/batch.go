// Package batch provides a generic batch operation wrapper with size validation.
package batch

import (
	"fmt"

	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// MaxSize is the maximum number of items allowed in a single batch operation.
const MaxSize = 50

// Batch is a generic wrapper for a batch of items.
type Batch[T any] struct {
	Items []T
}

// NewBatch creates a new Batch, validating the item count.
func NewBatch[T any](items []T) (*Batch[T], error) {
	if err := ValidateSize(items); err != nil {
		return nil, err
	}
	return &Batch[T]{Items: items}, nil
}

// ErrBatchSizeExceeded is returned when the batch size exceeds MaxSize.
var ErrBatchSizeExceeded = errors.NewDomainError(errors.BatchSizeExceeded,
	"batch size exceeds maximum allowed (50)", nil)

// ValidateSize checks that the number of items does not exceed MaxSize.
// Returns nil for 0..MaxSize items, ErrBatchSizeExceeded otherwise.
func ValidateSize[T any](items []T) error {
	if len(items) > MaxSize {
		return errors.NewDomainError(errors.BatchSizeExceeded,
			fmt.Sprintf("batch size %d exceeds maximum allowed %d", len(items), MaxSize), nil).
			WithDetails(fmt.Sprintf("actual: %d", len(items))).
			WithDetails(fmt.Sprintf("max: %d", MaxSize))
	}
	return nil
}

// Len returns the number of items in the batch.
func (b *Batch[T]) Len() int {
	return len(b.Items)
}
