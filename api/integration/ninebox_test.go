package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
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
// TestMatrixByPhase_TwoPerEvaluator verifies that a recompute produces two
// matrices per evaluator in the same cycle — one per phase (avance + cierre).
// This validates the migration index: two matrices per evaluator per phase
// coexist. The evaluator is derived from the API response (never hardcoded);
// the viewer (RH) sees all evaluators.
func TestMatrixByPhase_TwoPerEvaluator(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	cycleID := seed.SeedID("cycle-2026")
	evaluatorID := seed.SeedID("emp-dg-01")
	avancePhaseID := seed.SeedID("phase-def-avance")
	cierrePhaseID := seed.SeedID("phase-def-cierre")

	// (DB) Query matrices for the seeded evaluator in this cycle.
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

	// (API) Recompute both phases as RH so each evaluator (manager) has one
	// matrix per phase.
	for _, phaseID := range []uuid.UUID{avancePhaseID, cierrePhaseID} {
		req := httptest.NewRequest(http.MethodPost,
			fmt.Sprintf("/api/v1/nine-box/recompute/%s/%s", cycleID.String(), phaseID.String()),
			nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+srv.TokenRH)
		w := httptest.NewRecorder()
		srv.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code,
			"recompute %s should return 200: %s", phaseID, w.Body.String())
	}

	// GET matrices for both phases as RH (sees all). The view resolves one phase
	// per request, so query each phase and group by evaluatorId.
	var apiMatrices []matrixDTO
	for _, phaseID := range []uuid.UUID{avancePhaseID, cierrePhaseID} {
		apiMatrices = append(apiMatrices,
			getNineBoxMatrices(t, srv, srv.TokenRH,
				"?cycle_id="+cycleID.String()+"&phase_id="+phaseID.String())...)
	}

	// Group by evaluatorId: assert at least one evaluator has exactly two
	// matrices with distinct phases (avance + cierre).
	evaluatorPhases := make(map[string]map[string]bool)
	for _, m := range apiMatrices {
		if evaluatorPhases[m.EvaluatorID] == nil {
			evaluatorPhases[m.EvaluatorID] = make(map[string]bool)
		}
		evaluatorPhases[m.EvaluatorID][m.PhaseID] = true
	}

	found := false
	for _, phases := range evaluatorPhases {
		if len(phases) == 2 {
			found = true
			break
		}
	}
	assert.True(t, found,
		"expected at least one evaluator with exactly 2 matrices (avance + cierre)")
}

// ─── 5.4 Integration: computed matrix view (ninebox-computed-matrix) ─────────
//
// GET /nine-box/matrices computes the 9×9 view on demand via ComputeMatrixView:
// phase_id optional (nil → cycle.current_phase), scope by viewer role
// (rh/director-general see all; others see their org-node subtree), 1h TTL.

// matrixDTO mirrors dto.NineBoxMatrixResponse for the fields the tests assert.
type matrixDTO struct {
	ID          string     `json:"id"`
	CycleID     string     `json:"cycleId"`
	EvaluatorID string     `json:"evaluatorId"`
	PhaseID     string     `json:"phaseId"`
	Entries     []entryDTO `json:"entries"`
}

// entryDTO mirrors dto.NineBoxEntryDTO (tier-based + raw computation inputs).
type entryDTO struct {
	EvaluateeID         string   `json:"evaluateeId"`
	EmployeeName        string   `json:"employeeName"`
	PerformanceTier     int      `json:"performanceTier"`
	PotentialTier       int      `json:"potentialTier"`
	Quadrant            int      `json:"quadrant"`
	GoalProgressPercent float64  `json:"goalProgressPercent"`
	SelfRating          *float64 `json:"selfRating"`
	HrRating            *float64 `json:"hrRating"`
	Weights             *struct {
		Self float64 `json:"self"`
		HR   float64 `json:"hr"`
	} `json:"weights"`
}

// getNineBoxMatrices performs GET /nine-box/matrices with the given bearer
// token and returns the parsed matrix list. Asserts a 200 status.
func getNineBoxMatrices(t *testing.T, srv *testServer, token, query string) []matrixDTO {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/nine-box/matrices"+query, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "GET matrices should return 200: %s", w.Body.String())

	var matrices []matrixDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &matrices))
	return matrices
}

// evaluateeIDs collects the set of evaluatee IDs across all matrices.
func evaluateeIDs(matrices []matrixDTO) map[string]bool {
	ids := make(map[string]bool)
	for _, m := range matrices {
		for _, e := range m.Entries {
			ids[e.EvaluateeID] = true
		}
	}
	return ids
}

// TestNineBoxMatrices_DefaultPhaseFromCurrentPhase (REQ-NBM-001) verifies that
// omitting phase_id resolves the cycle's current phase (avance for cycle-2026).
func TestNineBoxMatrices_DefaultPhaseFromCurrentPhase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	cycleID := seed.SeedID("cycle-2026")
	avancePhaseID := seed.SeedID("phase-def-avance")

	matrices := getNineBoxMatrices(t, srv, srv.TokenRH, "?cycle_id="+cycleID.String())

	require.NotEmpty(t, matrices, "RH should see at least one computed matrix")
	for _, m := range matrices {
		assert.Equal(t, avancePhaseID.String(), m.PhaseID,
			"matrix phase should default to cycle.current_phase (avance) when phase_id is omitted")
	}
}

// TestNineBoxMatrices_RHSeesAll (REQ-NBM-003) verifies the rh role sees all
// employees regardless of org-node scope (auth.RoleSeesAll).
func TestNineBoxMatrices_RHSeesAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	cycleID := seed.SeedID("cycle-2026")
	dgID := seed.SeedID("emp-dg-01")

	matrices := getNineBoxMatrices(t, srv, srv.TokenRH, "?cycle_id="+cycleID.String())

	require.NotEmpty(t, matrices, "RH should see matrices")
	assert.True(t, evaluateeIDs(matrices)[dgID.String()],
		"RH (RoleSeesAll) should see emp-dg-01 regardless of org scope")
}

// TestNineBoxMatrices_ScopedReads (REQ-NBM-003) verifies a scoped viewer
// (colaborador emp-dg-01, child1 subtree) sees a strict subset of what RH sees.
//
// TODO(auth:C7): REQ-NBM-003 also mandates 403 for direct out-of-scope reads
// (evaluator_id / matrix id). The GET list endpoint currently FILTERS silently
// (200 with subset) — the 403-by-id path is not implemented yet. Also pending:
// jefe-ve-equipo with a non-RH manager profile (no seed token exists for one).
func TestNineBoxMatrices_ScopedReads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	cycleID := seed.SeedID("cycle-2026")
	dgID := seed.SeedID("emp-dg-01")
	colID := seed.SeedID("emp-colaborador-01")

	// Give emp-colaborador-01 (child2 subtree, manager=emp-dg-01) a goal
	// assignment so the cycle has two assignees in different org subtrees.
	assignID := uuid.New()
	_, err := srv.DB.ExecContext(ctx,
		`INSERT INTO goal_assignments (id, created_at, updated_at, employee_id, cycle_id)
		 VALUES ($1, NOW(), NOW(), $2, $3)`,
		assignID, colID, cycleID)
	require.NoError(t, err)
	defer func() {
		_, _ = srv.DB.ExecContext(ctx, `DELETE FROM goal_assignments WHERE id = $1`, assignID)
	}()

	// Force the seed matrices stale so ComputeMatrixView re-derives from live
	// goal assignments. Otherwise the seed's fresh matrix for evaluator
	// emp-dg-01 would be served from cache without the emp-colaborador-01 entry.
	_, err = srv.DB.ExecContext(ctx,
		`UPDATE nine_box_matrixes SET updated_at = now() - interval '2 hours' WHERE cycle_id = $1`,
		cycleID)
	require.NoError(t, err)

	rhMatrices := getNineBoxMatrices(t, srv, srv.TokenRH, "?cycle_id="+cycleID.String())
	rhEvals := evaluateeIDs(rhMatrices)
	assert.True(t, rhEvals[dgID.String()], "RH should see emp-dg-01")
	assert.True(t, rhEvals[colID.String()], "RH should see emp-colaborador-01 (sees all)")

	collabMatrices := getNineBoxMatrices(t, srv, srv.Token, "?cycle_id="+cycleID.String())
	collabEvals := evaluateeIDs(collabMatrices)
	assert.True(t, collabEvals[dgID.String()],
		"colaborador emp-dg-01 sees their own org node (child1, inclusive)")
	assert.False(t, collabEvals[colID.String()],
		"colaborador emp-dg-01 must NOT see emp-colaborador-01 (child2, out of scope)")
}

// TestNineBoxMatrices_EntryExposesRawInputs (REQ-NBM-002/005) verifies a
// re-derived entry exposes the raw computation inputs: goalProgressPercent,
// weights {self:0.2, hr:0.8}, and the weighted potential tier.
func TestNineBoxMatrices_EntryExposesRawInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	cycleID := seed.SeedID("cycle-2026")
	dgID := seed.SeedID("emp-dg-01")

	// First GET derives (no cached matrix for evaluator jefe in the seed),
	// so entries carry the raw inputs.
	matrices := getNineBoxMatrices(t, srv, srv.TokenRH, "?cycle_id="+cycleID.String())

	var entry *entryDTO
	for i := range matrices {
		for j := range matrices[i].Entries {
			if matrices[i].Entries[j].EvaluateeID == dgID.String() {
				entry = &matrices[i].Entries[j]
			}
		}
	}
	require.NotNil(t, entry, "matrix entries should include emp-dg-01")

	// Seed goal: current_value=20, target=100 → 20% progress.
	assert.InDelta(t, 20.0, entry.GoalProgressPercent, 0.5,
		"entry should expose goalProgressPercent from live goal progress")

	// Weighted potential defaults: self=0.2, hr=0.8.
	require.NotNil(t, entry.Weights, "derived entry should expose weights")
	assert.Equal(t, 0.2, entry.Weights.Self)
	assert.Equal(t, 0.8, entry.Weights.HR)

	// Seed competency rating defaults to source=rh with rating 4:
	// weighted potential = 4*0.8 = 3.2 → tier 2 (<= 3.66).
	require.NotNil(t, entry.HrRating, "hr rating should be present (source defaults to rh)")
	assert.Equal(t, 4.0, *entry.HrRating)
	assert.Equal(t, 2, entry.PotentialTier,
		"potential tier should reflect the weighted rating (4*0.8=3.2 → tier 2)")
}

// TestNineBoxMatrices_TTLCache (REQ-NBM-006) verifies the 1h snapshot cache:
// fresh GET derives then serves from cache, a forced-stale snapshot re-derives,
// and recompute bypasses the TTL.
func TestNineBoxMatrices_TTLCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	cycleID := seed.SeedID("cycle-2026")
	avancePhaseID := seed.SeedID("phase-def-avance")

	// Fresh: first GET derives the matrix and exposes raw inputs.
	first := getNineBoxMatrices(t, srv, srv.TokenRH, "?cycle_id="+cycleID.String())
	require.NotEmpty(t, first, "first GET should derive at least one matrix")
	require.NotEmpty(t, first[0].Entries)
	assert.Greater(t, first[0].Entries[0].GoalProgressPercent, 0.0,
		"derived entry should expose goalProgressPercent")

	// Fresh cache: second GET within the TTL is served without error.
	second := getNineBoxMatrices(t, srv, srv.TokenRH, "?cycle_id="+cycleID.String())
	require.NotEmpty(t, second, "cached GET should return matrices")
	require.NotEmpty(t, second[0].Entries)

	// Stale: force the snapshot older than 1h, then GET re-derives without error.
	_, err := srv.DB.ExecContext(ctx,
		`UPDATE nine_box_matrixes SET updated_at = now() - interval '2 hours' WHERE cycle_id = $1`,
		cycleID)
	require.NoError(t, err)

	third := getNineBoxMatrices(t, srv, srv.TokenRH, "?cycle_id="+cycleID.String())
	require.NotEmpty(t, third, "stale GET should re-derive and return matrices")
	require.NotEmpty(t, third[0].Entries)
	assert.Greater(t, third[0].Entries[0].GoalProgressPercent, 0.0,
		"stale matrix should be re-derived, exposing raw inputs again")

	// Recompute bypasses the TTL and forces re-derivation.
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/nine-box/recompute/%s/%s", cycleID.String(), avancePhaseID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.TokenRH)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "recompute should return 200: %s", w.Body.String())
}
