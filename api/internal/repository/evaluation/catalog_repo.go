package evaluation

import (
	"context"
	"sync"
	"time"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/nineboxquadrant"
	"github.com/sed-evaluacion-desempeno/api/internal/nineboxscale"
)

// cacheEntry holds cached data with its expiration time.
type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

// CatalogRepo provides read-only access to NineBoxQuadrant and NineBoxScale
// catalog tables with a simple in-memory cache (1h TTL).
type CatalogRepo struct {
	client *internal.Client

	mu       sync.RWMutex
	quadrants *cacheEntry
	scales    *cacheEntry
}

// NewCatalogRepo creates a new CatalogRepo.
func NewCatalogRepo(client *internal.Client) *CatalogRepo {
	return &CatalogRepo{client: client}
}

// GetQuadrants returns all 9 quadrant definitions with caching.
func (r *CatalogRepo) GetQuadrants(ctx context.Context) ([]*internal.NineBoxQuadrant, error) {
	r.mu.RLock()
	if r.quadrants != nil && time.Now().Before(r.quadrants.expiresAt) {
		data := r.quadrants.data.([]*internal.NineBoxQuadrant)
		r.mu.RUnlock()
		return data, nil
	}
	r.mu.RUnlock()

	results, err := r.client.NineBoxQuadrant.Query().
		Order(nineboxquadrant.ByQuadrant()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if results == nil {
		results = []*internal.NineBoxQuadrant{}
	}

	r.mu.Lock()
	r.quadrants = &cacheEntry{
		data:      results,
		expiresAt: time.Now().Add(1 * time.Hour),
	}
	r.mu.Unlock()

	return results, nil
}

// GetQuadrantByNumber returns a single quadrant by its number (1-9).
func (r *CatalogRepo) GetQuadrantByNumber(ctx context.Context, quadrant int) (*internal.NineBoxQuadrant, error) {
	all, err := r.GetQuadrants(ctx)
	if err != nil {
		return nil, err
	}
	for _, q := range all {
		if q.Quadrant == quadrant {
			return q, nil
		}
	}
	return nil, nil
}

// GetQuadrantByScores maps performance and potential scores to a quadrant
// and returns its metadata.
func (r *CatalogRepo) GetQuadrantByScores(ctx context.Context, perf, pot int) (*internal.NineBoxQuadrant, error) {
	q := computeQuadrantFromScores(perf, pot)
	if q == 0 {
		return nil, nil
	}
	return r.GetQuadrantByNumber(ctx, q)
}

// computeQuadrantFromScores is a local pure function that computes the quadrant.
// Duplicated from pkg/quadrant to avoid import cycle or dependency.
func computeQuadrantFromScores(performance, potential int) int {
	if performance < 1 || performance > 9 || potential < 1 || potential > 9 {
		return 0
	}
	perfTier := scoreTier(performance)
	potTier := scoreTier(potential)
	return (potTier-1)*3 + perfTier
}

func scoreTier(score int) int {
	switch {
	case score <= 3:
		return 1
	case score <= 6:
		return 2
	default:
		return 3
	}
}

// GetScales returns all 18 scale rows with caching.
func (r *CatalogRepo) GetScales(ctx context.Context) ([]*internal.NineBoxScale, error) {
	r.mu.RLock()
	if r.scales != nil && time.Now().Before(r.scales.expiresAt) {
		data := r.scales.data.([]*internal.NineBoxScale)
		r.mu.RUnlock()
		return data, nil
	}
	r.mu.RUnlock()

	results, err := r.client.NineBoxScale.Query().
		Order(nineboxscale.ByAxis(), nineboxscale.ByLevel()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if results == nil {
		results = []*internal.NineBoxScale{}
	}

	r.mu.Lock()
	r.scales = &cacheEntry{
		data:      results,
		expiresAt: time.Now().Add(1 * time.Hour),
	}
	r.mu.Unlock()

	return results, nil
}

// GetScalesByAxis filters scales by axis ("performance" or "potential").
func (r *CatalogRepo) GetScalesByAxis(ctx context.Context, axis string) ([]*internal.NineBoxScale, error) {
	all, err := r.GetScales(ctx)
	if err != nil {
		return nil, err
	}
	var filtered []*internal.NineBoxScale
	for _, s := range all {
		if string(s.Axis) == axis {
			filtered = append(filtered, s)
		}
	}
	if filtered == nil {
		return []*internal.NineBoxScale{}, nil
	}
	return filtered, nil
}

// GetScaleByAxisAndLevel returns a single scale definition for the given axis and level.
func (r *CatalogRepo) GetScaleByAxisAndLevel(ctx context.Context, axis string, level int) (*internal.NineBoxScale, error) {
	all, err := r.GetScales(ctx)
	if err != nil {
		return nil, err
	}
	for _, s := range all {
		if string(s.Axis) == axis && s.Level == level {
			return s, nil
		}
	}
	return nil, nil
}
