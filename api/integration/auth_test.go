package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUnauthenticatedRequestsReturn401 verifies that protected endpoints
// return 401 Unauthorized with a JSON error body when no session token
// or Authorization header is provided.
//
// Every endpoint behind RequireAuth (wired in PR4) should return 401
// before reaching any handler logic.
func TestUnauthenticatedRequestsReturn401(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "GET cycles without auth",
			method: "GET",
			path:   "/api/v1/cycles",
		},
		{
			name:   "POST cycles without auth",
			method: "POST",
			path:   "/api/v1/cycles",
		},
		{
			name:   "GET evaluations without auth",
			method: "GET",
			path:   "/api/v1/evaluations?cycle_id=11111111-1111-1111-1111-111111111111",
		},
		{
			name:   "GET org-trees without auth",
			method: "GET",
			path:   "/api/v1/org-trees",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			srv.Router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", w.Code)
			}

			// Verify response body is valid JSON with error.message
			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("response body is not valid JSON: %v\nBody: %s",
					err, w.Body.String())
			}

			errorObj, ok := resp["error"].(map[string]interface{})
			if !ok {
				t.Fatal("expected 'error' key in JSON response")
			}

			msg, ok := errorObj["message"].(string)
			if !ok || msg == "" {
				t.Error("expected non-empty 'error.message' in response")
			}
		})
	}
}
