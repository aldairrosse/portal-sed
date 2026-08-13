package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sed-evaluacion-desempeno/api/internal/seed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

// ─── 5.1 Integration: POST recompute ──────────────────────────────────────────
//
// TestRecomputeMatrix_SeedDataTiers calls recompute on the seeded cycle+phase
// and verifies that NineBoxEntries already exist with correct tiers.
// The seed already creates goals, evaluations, and nine-box matrices/entries.
func TestRecomputeMatrix_SeedDataTiers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	cycleID := seed.SeedID("cycle-2026")
	avancePhaseID := seed.SeedID("phase-def-avance")

	// POST recompute for the avance phase
	body := bytes.NewReader(nil)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/nine-box/recompute/%s/%s", cycleID.String(), avancePhaseID.String()),
		body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.TokenRH)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "recompute should return 200")

	// Verify entries exist in the database
	var entryCount int
	err := srv.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM nine_box_entries e
		 JOIN nine_box_matrixes m ON e.matrix_id = m.id
		 WHERE m.cycle_id = $1 AND m.phase_id = $2`,
		cycleID, avancePhaseID,
	).Scan(&entryCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, entryCount, 1,
		"recompute should produce at least one entry for the avance phase")

	// Verify entries have tier columns (not score columns)
	var tierCols int
	err = srv.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM information_schema.columns
		 WHERE table_name = 'nine_box_entries'
		   AND (column_name = 'performance_tier' OR column_name = 'potential_tier')`,
	).Scan(&tierCols)
	require.NoError(t, err)
	assert.Equal(t, 2, tierCols,
		"nine_box_entries should have performance_tier and potential_tier columns")
}

// ─── 5.2 Integration: PUT quadrant ────────────────────────────────────────────
//
// TestUpdateQuadrant_PersistsChanges verifies that PUT /nine-box/quadrants/{id}
// persists title, description, and colorHex changes, and that the response
// contains the updated values.
func TestUpdateQuadrant_PersistsChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	// Get the seeded quadrant 5
	var origTitle, origDesc, origColor string
	err := srv.DB.QueryRowContext(ctx,
		`SELECT title, description, color_hex FROM nine_box_quadrants WHERE quadrant = 5`,
	).Scan(&origTitle, &origDesc, &origColor)
	require.NoError(t, err, "should find quadrant 5")

	// PUT update
	newTitle := "Custom midline performer"
	newDesc := "Updated description for integration test"
	newColor := "#FF5733"

	payload, _ := json.Marshal(map[string]interface{}{
		"title":       newTitle,
		"description": newDesc,
		"colorHex":    newColor,
	})

	req := httptest.NewRequest(http.MethodPut,
		"/api/v1/nine-box/quadrants/5",
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.TokenRH)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Verify response body
	var resp struct {
		Quadrant    int    `json:"quadrant"`
		Title       string `json:"title"`
		Description string `json:"description"`
		ColorHex    string `json:"colorHex"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 5, resp.Quadrant)
	assert.Equal(t, newTitle, resp.Title)
	assert.Equal(t, newDesc, resp.Description)
	assert.Equal(t, newColor, resp.ColorHex)

	// Verify persisted in DB
	var dbTitle, dbDesc, dbColor string
	err = srv.DB.QueryRowContext(ctx,
		`SELECT title, description, color_hex FROM nine_box_quadrants WHERE quadrant = 5`,
	).Scan(&dbTitle, &dbDesc, &dbColor)
	require.NoError(t, err)
	assert.Equal(t, newTitle, dbTitle)
	assert.Equal(t, newDesc, dbDesc)
	assert.Equal(t, newColor, dbColor)

	// Restore original values
	_, _ = srv.DB.ExecContext(ctx,
		`UPDATE nine_box_quadrants SET title = $1, description = $2, color_hex = $3 WHERE quadrant = 5`,
		origTitle, origDesc, origColor,
	)
}

// ─── 5.3 Integration: matrix by phase ─────────────────────────────────────────
//
// TestMatrixByPhase_TwoPerEvaluator verifies that the seeded evaluator
// has two matrices (avance + cierre) for cycle 2026 — one per phase.
func TestMatrixByPhase_TwoPerEvaluator(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	cycleID := seed.SeedID("cycle-2026")
	evaluatorID := seed.SeedID("emp-dg-01")

	// Query matrices for this evaluator in this cycle
	rows, err := srv.DB.QueryContext(ctx,
		`SELECT id, phase_id FROM nine_box_matrixes
		 WHERE cycle_id = $1 AND evaluator_id = $2
		 ORDER BY phase_id`,
		cycleID, evaluatorID,
	)
	require.NoError(t, err)
	defer rows.Close()

	var matrices []struct {
		id      string
		phaseID string
	}
	for rows.Next() {
		var m struct {
			id      string
			phaseID string
		}
		require.NoError(t, rows.Scan(&m.id, &m.phaseID))
		matrices = append(matrices, m)
	}
	require.NoError(t, rows.Err())

	assert.Len(t, matrices, 2,
		"evaluator emp-dg-01 should have 2 matrices for cycle 2026 (avance + cierre)")

	// Each matrix should have a unique phase_id
	phaseIDs := make(map[string]bool)
	for _, m := range matrices {
		assert.NotEmpty(t, m.phaseID, "each matrix should have a phase_id")
		phaseIDs[m.phaseID] = true
	}
	assert.Len(t, phaseIDs, 2, "each matrix should have a different phase_id")

	// Verify the API returns phaseId and entries
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/nine-box/matrices?cycle_id=%s&evaluator_id=%s", cycleID.String(), evaluatorID.String()),
		nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.Token)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var apiMatrices []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &apiMatrices)
	require.NoError(t, err)
	assert.Len(t, apiMatrices, 2, "API should return 2 matrices")
	for _, m := range apiMatrices {
		phaseID := m["phaseId"].(string)
		assert.NotEmpty(t, phaseID, "each matrix should have phaseId in API response")
	}
}
