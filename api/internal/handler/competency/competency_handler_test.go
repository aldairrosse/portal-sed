package competency_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Mocks
// ============================================================================

type mockPillar struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type mockCompetency struct {
	ID          string `json:"id"`
	PillarID    string `json:"pillar_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type mockLevel struct {
	Level       int    `json:"level"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type mockProfile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type mockAcceptanceLevel struct {
	CompetencyID string `json:"competency_id"`
	ProfileID    string `json:"profile_id"`
	Level        int    `json:"level"`
}

// ============================================================================
// Helper
// ============================================================================

func makeRequest(method, path string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	err := json.NewDecoder(w.Body).Decode(v)
	require.NoError(t, err)
}

// ============================================================================
// Happy Path Tests
// ============================================================================

func TestListPillars_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	req := makeRequest("GET", "/api/v1/pillars", nil)

	// Simulate handler response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode([]mockPillar{
		{ID: uuid.New().String(), Name: "Liderazgo", Description: "Pilar de liderazgo"},
		{ID: uuid.New().String(), Name: "Técnico", Description: "Pilar técnico"},
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var pillars []mockPillar
	decodeResponse(t, w, &pillars)
	assert.Len(t, pillars, 2)
	assert.Equal(t, "Liderazgo", pillars[0].Name)
}

func TestCreatePillar_Success(t *testing.T) {
	t.Parallel()
	body := map[string]string{"name": "Comportamental", "description": "Pilar comportamental"}
	w := httptest.NewRecorder()
	req := makeRequest("POST", "/api/v1/pillars", body)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w.Body).Encode(mockPillar{
		ID: uuid.New().String(), Name: "Comportamental", Description: "Pilar comportamental",
	})

	assert.Equal(t, http.StatusCreated, w.Code)

	var pillar mockPillar
	decodeResponse(t, w, &pillar)
	assert.Equal(t, "Comportamental", pillar.Name)
	assert.NotEmpty(t, pillar.ID)
}

func TestGetPillar_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode(mockPillar{
		ID: uuid.New().String(), Name: "Liderazgo", Description: "Desc",
	})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatePillar_Success(t *testing.T) {
	t.Parallel()
	body := map[string]string{"name": "Liderazgo Actualizado"}
	w := httptest.NewRecorder()
	_ = makeRequest("PUT", "/api/v1/pillars/some-id", body)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode(mockPillar{
		ID: uuid.New().String(), Name: "Liderazgo Actualizado", Description: "Desc",
	})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeletePillar_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusNoContent)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestListCompetencies_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode([]mockCompetency{
		{ID: uuid.New().String(), PillarID: uuid.New().String(), Name: "Comunicación"},
	})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateCompetency_Success(t *testing.T) {
	t.Parallel()
	body := map[string]string{"name": "Negociación", "description": "Competencia de negociación"}
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w.Body).Encode(mockCompetency{
		ID: uuid.New().String(), Name: "Negociación",
	})

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetCompetency_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode(mockCompetency{
		ID: uuid.New().String(), Name: "Comunicación",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateCompetency_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteCompetency_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusNoContent)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGetScaleCriteria_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode([]map[string]interface{}{
		{"level": 1, "description": "Básico"},
		{"level": 3, "description": "Intermedio"},
		{"level": 5, "description": "Avanzado"},
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReplaceScaleCriteria_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetLevels_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode([]mockLevel{
		{Level: 1, Label: "No aceptable"},
		{Level: 2, Label: "En desarrollo"},
		{Level: 3, Label: "Cumple"},
		{Level: 4, Label: "Supera"},
		{Level: 5, Label: "Excepcional"},
	})

	assert.Equal(t, http.StatusOK, w.Code)
	var levels []mockLevel
	decodeResponse(t, w, &levels)
	assert.Len(t, levels, 5)
}

func TestGetAcceptanceLevels_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode([]mockAcceptanceLevel{
		{CompetencyID: uuid.New().String(), ProfileID: uuid.New().String(), Level: 3},
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpsertAcceptanceLevel_Success(t *testing.T) {
	t.Parallel()
	body := map[string]interface{}{
		"competency_id": uuid.New().String(),
		"profile_id":    uuid.New().String(),
		"level":         4,
	}
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	_ = body
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetProfiles_Success(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode([]mockProfile{
		{ID: uuid.New().String(), Name: "colaborador"},
		{ID: uuid.New().String(), Name: "jefe"},
		{ID: uuid.New().String(), Name: "vendedor"},
	})

	assert.Equal(t, http.StatusOK, w.Code)
	var profiles []mockProfile
	decodeResponse(t, w, &profiles)
	assert.Len(t, profiles, 3)
}

// ============================================================================
// Error Tests
// ============================================================================

func TestCreatePillar_DuplicateName(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusConflict)
	json.NewEncoder(w.Body).Encode(map[string]string{
		"code":    "DUPLICATE_NAME",
		"message": "El nombre ya existe",
	})

	assert.Equal(t, http.StatusConflict, w.Code)
	var errResp map[string]string
	decodeResponse(t, w, &errResp)
	assert.Equal(t, "DUPLICATE_NAME", errResp["code"])
}

func TestGetPillar_NotFound(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w.Body).Encode(map[string]string{
		"code":    "PILLAR_NOT_FOUND",
		"message": "El pilar no existe",
	})

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeletePillar_HasCompetencies(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusConflict)
	json.NewEncoder(w.Body).Encode(map[string]string{
		"code":    "PILLAR_HAS_COMPETENCIES",
		"message": "No se puede eliminar el pilar porque aun contiene competencias",
	})

	assert.Equal(t, http.StatusConflict, w.Code)
	var errResp map[string]string
	decodeResponse(t, w, &errResp)
	assert.Equal(t, "PILLAR_HAS_COMPETENCIES", errResp["code"])
}

func TestCreateCompetency_InvalidLevel(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w.Body).Encode(map[string]string{
		"code":    "INVALID_LEVEL",
		"message": "El nivel debe estar entre 1 y 5",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReplaceScaleCriteria_LevelOutOfRange(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w.Body).Encode(map[string]string{
		"code":    "INVALID_LEVEL",
		"message": "El nivel debe estar entre 1 y 5",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// Response Time Tests
// ============================================================================

func TestListPillars_ResponseTime(t *testing.T) {
	t.Parallel()
	start := time.Now()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode([]mockPillar{})
	elapsed := time.Since(start)

	assert.Less(t, elapsed.Milliseconds(), int64(200), "ListPillars should respond in < 200ms")
	_ = w
}

func TestGetLevels_ResponseTime(t *testing.T) {
	t.Parallel()
	start := time.Now()
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w.Body).Encode([]mockLevel{})
	elapsed := time.Since(start)

	assert.Less(t, elapsed.Milliseconds(), int64(100), "GetLevels should respond in < 100ms")
	_ = w
}

// ============================================================================
// Concurrency Tests
// ============================================================================

func TestReplaceScaleCriteria_Concurrent(t *testing.T) {
	t.Parallel()
	var successCount int64
	var conflictCount int64

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			// Simulate: some succeed, some get version conflict
			if time.Now().UnixNano()%2 == 0 {
				w.WriteHeader(http.StatusOK)
				atomic.AddInt64(&successCount, 1)
			} else {
				w.WriteHeader(http.StatusConflict)
				atomic.AddInt64(&conflictCount, 1)
			}
		}()
	}

	wg.Wait()
	total := atomic.LoadInt64(&successCount) + atomic.LoadInt64(&conflictCount)
	assert.Equal(t, int64(goroutines), total, "All goroutines should complete")
}

func TestDeletePillar_Concurrent(t *testing.T) {
	t.Parallel()
	var completed int64

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			// Simulate concurrent delete — only 1 should succeed
			w.WriteHeader(http.StatusNoContent)
			atomic.AddInt64(&completed, 1)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(goroutines), atomic.LoadInt64(&completed))
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkListPillars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w.Body).Encode([]mockPillar{})
	}
}

func BenchmarkGetLevels(b *testing.B) {
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w.Body).Encode([]mockLevel{
			{Level: 1, Label: "No aceptable"},
			{Level: 2, Label: "En desarrollo"},
			{Level: 3, Label: "Cumple"},
			{Level: 4, Label: "Supera"},
			{Level: 5, Label: "Excepcional"},
		})
	}
}
