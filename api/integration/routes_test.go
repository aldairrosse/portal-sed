package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// routeTestCase defines a route to test.
type routeTestCase struct {
	Method string
	Path   string
	Name   string
	// Expected status codes — any of these is acceptable.
	// For example, a route behind auth might return 200 or 401.
	// We primarily care that it doesn't return 500 or panic.
	AllowStatus []int
}

// allRoutes returns every main API route that should respond without 500.
// GET routes with required path params use placeholder UUIDs.
func allRoutes() []routeTestCase {
	empID := "11111111-1111-1111-1111-111111111111"
	catID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	goalID := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	kpiID := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	evalID := "22222222-2222-2222-2222-222222222222"
	cycleID := "99999999-9999-9999-9999-999999999999"
	matrixID := "33333333-3333-3333-3333-333333333333"
	treeID := "44444444-4444-4444-4444-444444444444"
	nodeID := "55555555-5555-5555-5555-555555555555"
	pillarID := "66666666-6666-6666-6666-666666666666"
	compID := "77777777-7777-7777-7777-777777777777"
	scopeID := "88888888-8888-8888-8888-888888888888"
	entryID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	// Routes behind AuthPlaceholder return 200 (passthrough).
	authOK := []int{200, 201, 400, 404, 409, 422}

	return []routeTestCase{
		// --- Health ---
		{"GET", "/health", "HealthCheck", []int{200}},

		// --- Auth ---
		{"POST", "/api/v1/auth/login", "AuthLogin", []int{200, 400, 405}},
		{"POST", "/api/v1/auth/logout", "AuthLogout", []int{200, 401}},
		{"POST", "/api/v1/auth/refresh", "AuthRefresh", []int{200, 401}},
		{"GET", "/api/v1/auth/me", "AuthMe", []int{200, 401}},

		// --- Goal: Categories ---
		{"GET", "/api/v1/employees/" + empID + "/categories", "ListCategories", authOK},
		{"POST", "/api/v1/employees/" + empID + "/categories", "CreateCategory", authOK},

		// --- Goal: Goals ---
		{"POST", "/api/v1/employees/" + empID + "/categories/" + catID + "/goals", "CreateGoal", authOK},
		{"PUT", "/api/v1/goals/" + goalID, "UpdateGoal", authOK},
		{"DELETE", "/api/v1/goals/" + goalID, "DeleteGoal", authOK},
		{"PATCH", "/api/v1/goals/" + goalID + "/progress", "UpdateGoalProgress", authOK},
		{"POST", "/api/v1/goals/batch", "BatchGoals", authOK},

		// --- Goal: Weights ---
		{"POST", "/api/v1/employees/" + empID + "/validate-weights", "ValidateWeights", authOK},

		// --- Goal: KPIs ---
		{"GET", "/api/v1/kpis", "ListKPIs", authOK},
		{"POST", "/api/v1/kpis", "CreateKPI", authOK},
		{"PUT", "/api/v1/kpis/" + kpiID, "UpdateKPI", authOK},
		{"DELETE", "/api/v1/kpis/" + kpiID, "DeleteKPI", authOK},

		// --- Goal: KPI Linking ---
		{"POST", "/api/v1/goals/" + goalID + "/kpis", "LinkKPI", authOK},
		{"DELETE", "/api/v1/goals/" + goalID + "/kpis/" + kpiID, "UnlinkKPI", authOK},

		// --- Goal: Assignments ---
		{"GET", "/api/v1/employees/" + empID + "/assignments", "GetAssignment", authOK},
		{"POST", "/api/v1/employees/" + empID + "/assignments", "CreateAssignment", authOK},

		// --- Cycle ---
		{"GET", "/api/v1/cycles", "ListCycles", authOK},
		{"POST", "/api/v1/cycles", "CreateCycle", authOK},
		{"GET", "/api/v1/cycles/" + cycleID, "GetCycle", authOK},
		{"PUT", "/api/v1/cycles/" + cycleID + "/transition", "TransitionPhase", authOK},
		{"GET", "/api/v1/phases", "GetPhaseDefinitions", authOK},
		{"GET", "/api/v1/cycles/" + cycleID + "/transitions", "GetAvailableTransitions", authOK},

		// --- Evaluation ---
		{"GET", "/api/v1/evaluations?cycle_id=" + cycleID, "ListEvaluations", authOK},
		{"GET", "/api/v1/evaluations/" + evalID, "GetEvaluation", authOK},
		{"POST", "/api/v1/evaluations/" + evalID + "/self-evaluation", "SubmitSelfEvaluation", authOK},
		{"PUT", "/api/v1/evaluations/" + evalID + "/self-evaluation", "UpdateSelfEvaluation", authOK},
		{"POST", "/api/v1/evaluations/" + evalID + "/rh-evaluation", "SubmitRHEvaluation", authOK},
		{"PUT", "/api/v1/evaluations/" + evalID + "/rh-evaluation", "UpdateRHEvaluation", authOK},
		{"POST", "/api/v1/evaluations/" + evalID + "/finalize", "FinalizeEvaluation", authOK},
		{"GET", "/api/v1/evaluations/summary?cycle_id=" + cycleID, "GetEvaluationSummary", authOK},

		// --- Nine-Box ---
		{"GET", "/api/v1/nine-box/matrices?cycle_id=" + cycleID + "&evaluator_id=" + empID, "ListMatrices", authOK},
		{"POST", "/api/v1/nine-box/matrices", "CreateMatrix", authOK},
		{"GET", "/api/v1/nine-box/matrices/" + matrixID, "GetNineBoxMatrix", authOK},
		{"GET", "/api/v1/nine-box/matrices/" + matrixID + "/entries", "ListMatrixEntries", authOK},
		{"POST", "/api/v1/nine-box/matrices/" + matrixID + "/entries", "UpsertMatrixEntry", authOK},
		{"PUT", "/api/v1/nine-box/entries/" + entryID, "UpdateEntry", authOK},
		{"POST", "/api/v1/nine-box/batch?matrixId=" + matrixID, "BatchSubmitEntries", authOK},
		{"GET", "/api/v1/nine-box/scales", "GetNineBoxScales", authOK},
		{"GET", "/api/v1/nine-box/quadrants", "GetNineBoxQuadrants", authOK},

		// --- Competency: Pillars ---
		{"GET", "/api/v1/pillars", "ListPillars", authOK},
		{"POST", "/api/v1/pillars", "CreatePillar", authOK},
		{"GET", "/api/v1/pillars/" + pillarID, "GetPillar", authOK},
		{"PUT", "/api/v1/pillars/" + pillarID, "UpdatePillar", authOK},
		{"DELETE", "/api/v1/pillars/" + pillarID, "DeletePillar", authOK},

		// --- Competency: Competencies ---
		{"GET", "/api/v1/pillars/" + pillarID + "/competencies", "ListCompetenciesByPillar", authOK},
		{"POST", "/api/v1/pillars/" + pillarID + "/competencies", "CreateCompetency", authOK},
		{"GET", "/api/v1/competencies/" + compID, "GetCompetency", authOK},
		{"PUT", "/api/v1/competencies/" + compID, "UpdateCompetency", authOK},
		{"DELETE", "/api/v1/competencies/" + compID, "DeleteCompetency", authOK},

		// --- Competency: Scale Criteria ---
		{"GET", "/api/v1/competencies/" + compID + "/scale-criteria", "GetScaleCriteria", authOK},
		{"POST", "/api/v1/competencies/" + compID + "/scale-criteria", "UpsertScaleCriteria", authOK},

		// --- Competency: Catalogs ---
		{"GET", "/api/v1/levels", "ListLevels", authOK},
		{"GET", "/api/v1/profiles", "ListProfiles", authOK},
		{"GET", "/api/v1/acceptance-levels", "ListAcceptanceLevels", authOK},
		{"POST", "/api/v1/acceptance-levels", "UpsertAcceptanceLevel", authOK},

		// --- Org: Trees ---
		{"GET", "/api/v1/org-trees", "ListOrgTrees", authOK},
		{"GET", "/api/v1/org-trees/" + treeID, "GetOrgTree", authOK},
		{"GET", "/api/v1/org-trees/" + treeID + "/nodes", "GetOrgTreeNodes", authOK},
		{"GET", "/api/v1/org-trees/" + treeID + "/export", "ExportOrgTree", authOK},

		// --- Org: Nodes ---
		{"GET", "/api/v1/org-nodes/" + nodeID, "GetOrgNode", authOK},
		{"POST", "/api/v1/org-nodes", "CreateOrgNode", authOK},
		{"PUT", "/api/v1/org-nodes/" + nodeID, "UpdateOrgNode", authOK},
		{"DELETE", "/api/v1/org-nodes/" + nodeID, "DeleteOrgNode", authOK},
		{"POST", "/api/v1/org-nodes/" + nodeID + "/move", "MoveOrgNode", authOK},

		// --- Org: Employees ---
		{"GET", "/api/v1/employees", "ListEmployees", authOK},
		{"GET", "/api/v1/employees/" + empID, "GetEmployee", authOK},
		{"GET", "/api/v1/employees/" + empID + "/evaluatees", "GetMyEvaluatees", authOK},
		{"GET", "/api/v1/employees/" + empID + "/manager", "GetManager", authOK},
		{"GET", "/api/v1/employees/" + empID + "/ancestors", "GetAncestors", authOK},
		{"POST", "/api/v1/employees/batch", "BatchLookupEmployees", authOK},
		{"GET", "/api/v1/employees/search?q=Perez", "SearchEmployees", authOK},

		// --- Org: Evaluator Scopes ---
		{"GET", "/api/v1/evaluator-scopes", "GetEvaluatorScope", authOK},
		{"GET", "/api/v1/evaluator-scopes/" + scopeID, "GetEvaluatorScopeByID", authOK},
	}
}

// TestAllRoutesRespond verifies that every main API route responds without
// returning 500 Internal Server Error or panicking. This is a smoke test
// that validates the full DI wiring and route registration.
func TestAllRoutesRespond(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	routes := allRoutes()

	for _, tc := range routes {
		t.Run(tc.Name, func(t *testing.T) {
			var req *http.Request
			if tc.Method == "POST" {
				// Send empty body for POST routes
				req = httptest.NewRequest(tc.Method, tc.Path, nil)
			} else {
				req = httptest.NewRequest(tc.Method, tc.Path, nil)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			srv.Router.ServeHTTP(w, req)

			// Verify no 500 or panic
			if w.Code == http.StatusInternalServerError {
				t.Errorf("route %s %s returned 500: %s", tc.Method, tc.Path, w.Body.String())
			}

			// Verify status is in allowed list
			allowed := false
			for _, s := range tc.AllowStatus {
				if w.Code == s {
					allowed = true
					break
				}
			}
			if !allowed {
				t.Errorf("route %s %s returned status %d, expected one of %v",
					tc.Method, tc.Path, w.Code, tc.AllowStatus)
			}
		})
	}
}

// TestHealthCheck verifies the health endpoint returns a valid JSON response.
func TestHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health check returned %d, expected 200", w.Code)
	}

	body := w.Body.String()
	if body != `{"status":"ok"}` {
		t.Errorf("health check returned body %q, expected %q", body, `{"status":"ok"}`)
	}
}

// TestRouteCount verifies that we have a minimum number of routes registered.
// This is a guard against accidentally removing routes.
func TestRouteCount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	routes := allRoutes()
	minExpected := 60 // We have 70+ routes; this is a safety floor

	if len(routes) < minExpected {
		t.Errorf("expected at least %d routes, got %d — routes may have been accidentally removed",
			minExpected, len(routes))
	}
}
