package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/seed"
	svcgoal "github.com/sed-evaluacion-desempeno/api/internal/service/goal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── 6.1 PATCH /kpis/{id}/value ──────────────────────────────────────────────

// TestPatchKPIValue_HappyPath updates a seeded KPI's current_value and verifies
// the response contains the updated value.
func TestPatchKPIValue_HappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	kpiID := seed.SeedID("kpi-ventas-mensuales")
	newValue := 850000.0

	payload, _ := json.Marshal(map[string]float64{"current_value": newValue})
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/api/v1/kpis/%s/value", kpiID.String()),
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.Token)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "PATCH KPI value should return 200")

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "response should be valid JSON")

	assert.Equal(t, kpiID.String(), resp["id"], "should return the same KPI")
	assert.Equal(t, newValue, resp["current_value"], "current_value should be updated")
	assert.Equal(t, "ascendente", resp["direction"], "should preserve direction")
	assert.NotEmpty(t, resp["name"], "should include name")

	// Verify persisted in DB
	var dbValue float64
	err = srv.DB.QueryRowContext(ctx,
		`SELECT COALESCE(current_value, 0) FROM kp_is WHERE id = $1`, kpiID,
	).Scan(&dbValue)
	require.NoError(t, err)
	assert.Equal(t, newValue, dbValue, "current_value should be persisted in DB")
}

// TestPatchKPIValue_KPINotFound verifies 404 for a non-existent KPI.
func TestPatchKPIValue_KPINotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	fakeID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	payload, _ := json.Marshal(map[string]float64{"current_value": 100.0})
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/api/v1/kpis/%s/value", fakeID.String()),
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.Token)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "non-existent KPI should return 404")

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "response should be valid JSON")
	assert.Contains(t, resp, "error", "should have error object")
}

// TestPatchKPIValue_NegativeValue verifies 400 for negative current_value.
func TestPatchKPIValue_NegativeValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	kpiID := seed.SeedID("kpi-ventas-mensuales")
	payload, _ := json.Marshal(map[string]float64{"current_value": -1.0})
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/api/v1/kpis/%s/value", kpiID.String()),
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.Token)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "negative value should return 400")

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "response should be valid JSON")
	assert.Contains(t, resp, "error", "should have error object")
}

// ─── 6.2 GET /employees/{id}/score ───────────────────────────────────────────

// TestGetEmployeeScore_ReturnsScore verifies the score endpoint returns a number
// between 0-100 for an employee with categories and goals.
func TestGetEmployeeScore_ReturnsScore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	empID := seed.SeedID("emp-dg-01")

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/employees/%s/score", empID.String()),
		nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.Token)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "get score should return 200")

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "response should be valid JSON")

	score, ok := resp["score"].(float64)
	require.True(t, ok, "response should have 'score' key with a number")
	assert.GreaterOrEqual(t, score, 0.0, "score should be >= 0")
	assert.LessOrEqual(t, score, 100.0, "score should be <= 100")
}

// TestGetEmployeeScore_NoData returns 0 for employee without categories.
func TestGetEmployeeScore_NoData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	// emp-colaborador-01 exists but has no goal categories (only emp-dg-01 has categories in seed)
	empID := seed.SeedID("emp-colaborador-01")

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/employees/%s/score", empID.String()),
		nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.Token)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "get score should return 200")

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "response should be valid JSON")

	score, ok := resp["score"].(float64)
	require.True(t, ok, "response should have 'score' key")
	assert.Equal(t, 0.0, score, "employee without categories should have score 0")
}

// ─── 6.3 Create descendente goal ─────────────────────────────────────────────

// TestCreateDescendenteGoal_Success creates a goal with direction=descendente
// and baselineValue, and verifies it persists correctly.
func TestCreateDescendenteGoal_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	empID := seed.SeedID("emp-dg-01")
	catID := seed.SeedID("cat-ventas-finanzas")
	baselineVal := 200.0

	payload := map[string]interface{}{
		"name":           "Test descendente goal",
		"description":    "Integration test for descendente",
		"unit":           "numero",
		"weight":         10.0,
		"target_value":   50.0,
		"direction":      "descendente",
		"baseline_value": baselineVal,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/employees/%s/categories/%s/goals", empID.String(), catID.String()),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.Token)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code,
		"create descendente goal should return 201, got %d: %s", w.Code, w.Body.String())

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "response should be valid JSON")

	assert.Equal(t, "descendente", resp["direction"], "direction should be descendente")

	// baseline_value may be a float64 or nil - check it exists
	bv, ok := resp["baseline_value"].(float64)
	require.True(t, ok, "baseline_value should be a number in response")
	assert.Equal(t, baselineVal, bv, "baseline_value should match")

	// Verify persisted in DB
	goalID := resp["id"].(string)
	var dbDirection string
	var dbBaseline float64
	err = srv.DB.QueryRowContext(ctx,
		`SELECT direction, COALESCE(baseline_value, 0) FROM goals WHERE id = $1`, goalID,
	).Scan(&dbDirection, &dbBaseline)
	require.NoError(t, err)
	assert.Equal(t, "descendente", dbDirection)
	assert.Equal(t, baselineVal, dbBaseline)
}

// TestCreateDescendenteGoal_MissingBaseline verifies that creating a descendente
// goal without baselineValue returns 400.
func TestCreateDescendenteGoal_MissingBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	empID := seed.SeedID("emp-dg-01")
	catID := seed.SeedID("cat-ventas-finanzas")

	payload := map[string]interface{}{
		"name":         "Test descendente no baseline",
		"description":  "Should fail without baseline",
		"unit":         "numero",
		"weight":       10.0,
		"target_value": 50.0,
		"direction":    "descendente",
		// no baseline_value
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/employees/%s/categories/%s/goals", empID.String(), catID.String()),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.Token)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code,
		"descendente without baseline should return 400, got %d: %s", w.Code, w.Body.String())

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "response should be valid JSON")

	errorObj, ok := resp["error"].(map[string]interface{})
	require.True(t, ok, "should have error object")
	assert.Contains(t, errorObj["message"], "baseline_value",
		"error message should mention baseline_value")
}

// ─── 6.4 Block direction change in phase avance ───────────────────────────────

// TestUpdateGoal_BlocksDirectionChangeInAvance creates a goal in asignacion phase,
// then attempts to update its direction using a server setup that reports
// "avance" as the current phase. The update should be rejected with 403.
func TestUpdateGoal_BlocksDirectionChangeInAvance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Step 1: Create a goal using the default server (phase=asignacion)
	srv1 := setupTestServer(t)
	defer srv1.Clean()

	empID := seed.SeedID("emp-dg-01")
	catID := seed.SeedID("cat-ventas-finanzas")

	createPayload := map[string]interface{}{
		"name":         "Goal for avance test",
		"description":  "Will try to change direction in avance",
		"unit":         "numero",
		"weight":       10.0,
		"target_value": 100.0,
		"direction":    "ascendente",
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/employees/%s/categories/%s/goals", empID.String(), catID.String()),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv1.Token)
	w := httptest.NewRecorder()
	srv1.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "goal creation should succeed in asignacion")

	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)
	goalID := createResp["id"].(string)
	version := createResp["version"].(float64)

	// Step 2: Set up a server that reports "avance" phase
	srv2 := setupTestServerWithPhaseChecker(t, fixedPhaseChecker{phase: svcgoal.PhaseAvance})
	defer srv2.Clean()

	// Try to change direction (should be blocked in avance)
	updatePayload := map[string]interface{}{
		"name":         "Goal for avance test",
		"description":  "Trying to change direction",
		"unit":         "numero",
		"weight":       10.0,
		"target_value": 100.0,
		"direction":    "descendente",
		"baseline_value": 200.0,
		"version":      int(version),
	}
	body2, _ := json.Marshal(updatePayload)
	req2 := httptest.NewRequest(http.MethodPut,
		fmt.Sprintf("/api/v1/goals/%s", goalID),
		bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+srv2.Token)
	w2 := httptest.NewRecorder()
	srv2.Router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusForbidden, w2.Code,
		"changing direction in avance should return 403, got %d: %s", w2.Code, w2.Body.String())

	var errResp map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &errResp)
	require.NoError(t, err, "response should be valid JSON")
	errorObj, ok := errResp["error"].(map[string]interface{})
	require.True(t, ok, "should have error object")
	assert.Equal(t, "PHASE_RESTRICTED", errorObj["code"], "error code should be PHASE_RESTRICTED")
}

// TestUpdateGoal_BlocksBaselineChangeInAvance verifies that changing baseline_value
// is also rejected when the phase is avance.
func TestUpdateGoal_BlocksBaselineChangeInAvance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Step 1: Create a descendente goal in asignacion phase
	srv1 := setupTestServer(t)
	defer srv1.Clean()

	empID := seed.SeedID("emp-dg-01")
	catID := seed.SeedID("cat-ventas-finanzas")

	createPayload := map[string]interface{}{
		"name":           "Baseline change test",
		"description":    "Will try to change baseline in avance",
		"unit":           "numero",
		"weight":         10.0,
		"target_value":   50.0,
		"direction":      "descendente",
		"baseline_value": 200.0,
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/employees/%s/categories/%s/goals", empID.String(), catID.String()),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv1.Token)
	w := httptest.NewRecorder()
	srv1.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "goal creation should succeed")

	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)
	goalID := createResp["id"].(string)
	version := createResp["version"].(float64)

	// Step 2: Avance phase server
	srv2 := setupTestServerWithPhaseChecker(t, fixedPhaseChecker{phase: svcgoal.PhaseAvance})
	defer srv2.Clean()

	// Same goal data but with different baseline_value
	updatePayload := map[string]interface{}{
		"name":           "Baseline change test",
		"description":    "Trying to change baseline in avance",
		"unit":           "numero",
		"weight":         10.0,
		"target_value":   50.0,
		"direction":      "descendente",
		"baseline_value": 300.0, // changed!
		"version":        int(version),
	}
	body2, _ := json.Marshal(updatePayload)
	req2 := httptest.NewRequest(http.MethodPut,
		fmt.Sprintf("/api/v1/goals/%s", goalID),
		bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+srv2.Token)
	w2 := httptest.NewRecorder()
	srv2.Router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusForbidden, w2.Code,
		"changing baseline_value in avance should return 403, got %d: %s", w2.Code, w2.Body.String())
}
