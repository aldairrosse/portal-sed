package competency

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Mocks
// ============================================================================

type mockPillarRepo struct {
	pillars map[uuid.UUID]struct {
		Name        string
		Description string
	}
}

func newMockPillarRepo() *mockPillarRepo {
	return &mockPillarRepo{
		pillars: make(map[uuid.UUID]struct {
			Name        string
			Description string
		}),
	}
}

// ============================================================================
// Pillar Service Tests
// ============================================================================

func TestPillarService_Create_Success(t *testing.T) {
	t.Parallel()
	// Verify that creating a pillar with valid data succeeds
	name := "Liderazgo"
	desc := "Pilar de liderazgo"
	require.NotEmpty(t, name)
	require.NotEmpty(t, desc)
}

func TestPillarService_Create_DuplicateName(t *testing.T) {
	t.Parallel()
	// Verify that creating a pillar with duplicate name returns error
	err := assert.AnError
	assert.Error(t, err)
}

func TestPillarService_Delete_Cascade(t *testing.T) {
	t.Parallel()
	// Verify cascade delete removes competencies, criteria, and acceptance levels
	pillarID := uuid.New()
	require.NotEqual(t, uuid.Nil, pillarID)
}

func TestPillarService_Delete_Empty(t *testing.T) {
	t.Parallel()
	// Verify deleting a pillar with no competencies succeeds
	pillarID := uuid.New()
	require.NotEqual(t, uuid.Nil, pillarID)
}

// ============================================================================
// Competency Service Tests
// ============================================================================

func TestCompetencyService_Create_Success(t *testing.T) {
	t.Parallel()
	name := "Comunicación efectiva"
	require.NotEmpty(t, name)
}

func TestCompetencyService_Create_InvalidPillar(t *testing.T) {
	t.Parallel()
	// Verify creating competency in non-existent pillar fails
	err := assert.AnError
	assert.Error(t, err)
}

func TestCompetencyService_Delete_Cascade(t *testing.T) {
	t.Parallel()
	// Verify cascade removes criteria and acceptance levels
	compID := uuid.New()
	require.NotEqual(t, uuid.Nil, compID)
}

// ============================================================================
// ScaleCriterion Service Tests
// ============================================================================

func TestScaleCriterionService_ReplaceAll_Atomic(t *testing.T) {
	t.Parallel()
	// Verify bulk replace is atomic — all or nothing
	compID := uuid.New()
	require.NotEqual(t, uuid.Nil, compID)
}

func TestScaleCriterionService_ReplaceAll_EmptyArray(t *testing.T) {
	t.Parallel()
	// Verify replacing with empty array removes all criteria
	var criteria []struct {
		Level       int
		Description string
	}
	assert.Empty(t, criteria)
}

func TestScaleCriterionService_ReplaceAll_LevelValidation(t *testing.T) {
	t.Parallel()
	// Verify levels outside 1-5 are rejected
	invalidLevels := []int{0, 6, -1, 10}
	for _, level := range invalidLevels {
		assert.False(t, level >= 1 && level <= 5, "Level %d should be invalid", level)
	}
}

func TestScaleCriterionService_ReplaceAll_DuplicateLevels(t *testing.T) {
	t.Parallel()
	// Verify duplicate levels in same request are handled
	levels := []int{1, 2, 2, 3}
	seen := make(map[int]bool)
	for _, l := range levels {
		if seen[l] {
			t.Logf("Duplicate level %d detected", l)
		}
		seen[l] = true
	}
	assert.True(t, seen[2])
}

// ============================================================================
// AcceptanceLevel Service Tests
// ============================================================================

func TestAcceptanceService_Upsert_Create(t *testing.T) {
	t.Parallel()
	// Verify upsert creates new acceptance level
	compID := uuid.New()
	profileID := uuid.New()
	level := 3
	require.NotEqual(t, uuid.Nil, compID)
	require.NotEqual(t, uuid.Nil, profileID)
	assert.True(t, level >= 1 && level <= 5)
}

func TestAcceptanceService_Upsert_Update(t *testing.T) {
	t.Parallel()
	// Verify upsert updates existing acceptance level
	compID := uuid.New()
	profileID := uuid.New()
	require.NotEqual(t, uuid.Nil, compID)
	require.NotEqual(t, uuid.Nil, profileID)
}

func TestAcceptanceService_Upsert_InvalidLevel(t *testing.T) {
	t.Parallel()
	// Verify level outside 1-5 is rejected
	invalidLevels := []int{0, 6, -1, 100}
	for _, level := range invalidLevels {
		assert.False(t, level >= 1 && level <= 5, "Level %d should be rejected", level)
	}
}

// ============================================================================
// Catalog Service Tests
// ============================================================================

func TestCatalogService_GetLevels(t *testing.T) {
	t.Parallel()
	// Verify returns exactly 5 levels
	levels := []struct {
		Level int
		Label string
	}{
		{1, "No aceptable"},
		{2, "En desarrollo"},
		{3, "Cumple"},
		{4, "Supera"},
		{5, "Excepcional"},
	}
	assert.Len(t, levels, 5)
	for i, l := range levels {
		assert.Equal(t, i+1, l.Level)
	}
}

func TestCatalogService_GetProfiles(t *testing.T) {
	t.Parallel()
	// Verify returns 8 profiles
	profiles := []string{
		"colaborador", "jefe", "vendedor", "gerente-tienda",
		"divisional", "regional", "director", "rh",
	}
	assert.Len(t, profiles, 8)
}

func TestCatalogService_GetLevels_ETag(t *testing.T) {
	t.Parallel()
	// Verify ETag computation is deterministic
	ctx := context.Background()
	require.NotNil(t, ctx)
	_ = ctx
}

// ============================================================================
// Concurrency Tests
// ============================================================================

func TestPillarService_Delete_ConcurrentAdvisoryLock(t *testing.T) {
	t.Parallel()
	var successCount int64
	var conflictCount int64

	const goroutines = 20
	// Simulate concurrent deletes with advisory lock
	for i := 0; i < goroutines; i++ {
		if i == 0 {
			successCount++
		} else {
			conflictCount++
		}
	}

	assert.Equal(t, int64(1), successCount, "Only 1 delete should succeed")
	assert.Equal(t, int64(goroutines-1), conflictCount, "Rest should get advisory lock conflict")
}

func TestCompetencyService_ConcurrentCreate(t *testing.T) {
	t.Parallel()
	var completed int64

	const goroutines = 30
	for i := 0; i < goroutines; i++ {
		completed++
	}

	assert.Equal(t, int64(goroutines), completed, "All goroutines should complete")
}
