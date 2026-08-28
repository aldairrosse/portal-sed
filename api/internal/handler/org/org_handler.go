// Package org provides HTTP handlers for the organizational hierarchy API.
// Handlers are thin: they validate input, call the service layer, and format
// responses. Business logic lives in the service layer.
package org

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/dto/org"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/org"
)

// generateTraceID generates a short trace ID for error responses.
func generateTraceID() string {
	id := uuid.New().String()
	return id[:8] + "-" + id[9:13]
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handler/org: failed to encode JSON response: %v", err)
	}
}

// writeError writes a structured error response.
func writeError(w http.ResponseWriter, err error) {
	traceID := generateTraceID()
	status := errors.HTTPStatus(err)

	var de *errors.DomainError
	if errors.AsDomainError(err, &de) {
		writeJSON(w, status, errors.NewAPIErrorResponse(de, traceID))
		return
	}

	writeJSON(w, status, errors.NewAPIErrorResponse(
		errors.NewDomainError(errors.InvalidRequest, err.Error(), err),
		traceID,
	))
}

// OrgHandler holds HTTP handlers for all org hierarchy operations.
type OrgHandler struct {
	treeSvc      svc.OrgTreeService
	nodeSvc      svc.OrgNodeService
	employeeSvc  svc.EmployeeService
	evaluateeSvc svc.EvaluateeService
	metricsSvc   svc.MetricsService
}

// NewOrgHandler creates a new OrgHandler.
func NewOrgHandler(
	treeSvc svc.OrgTreeService,
	nodeSvc svc.OrgNodeService,
	employeeSvc svc.EmployeeService,
	evaluateeSvc svc.EvaluateeService,
	metricsSvc svc.MetricsService,
) *OrgHandler {
	return &OrgHandler{
		treeSvc:      treeSvc,
		nodeSvc:      nodeSvc,
		employeeSvc:  employeeSvc,
		evaluateeSvc: evaluateeSvc,
		metricsSvc:   metricsSvc,
	}
}

// ==================== Tree Endpoints ====================

// ListOrgTrees handles GET /api/v1/org-trees
func (h *OrgHandler) ListOrgTrees(w http.ResponseWriter, r *http.Request) {
	treeType := r.URL.Query().Get("type")

	result, err := h.treeSvc.GetTrees(r.Context(), treeType)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetOrgTree handles GET /api/v1/org-trees/{treeId}
func (h *OrgHandler) GetOrgTree(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "treeId")
	if treeID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "treeId path parameter is required", nil))
		return
	}

	if _, err := uuid.Parse(treeID); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "treeId must be a valid UUID v4", err))
		return
	}

	result, err := h.treeSvc.GetTree(r.Context(), treeID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetOrgTreeNodes handles GET /api/v1/org-trees/{treeId}/nodes
func (h *OrgHandler) GetOrgTreeNodes(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "treeId")
	if treeID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "treeId path parameter is required", nil))
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "flat"
	}
	if format != "flat" && format != "nested" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "format must be 'flat' or 'nested'", nil))
		return
	}

	depth := -1
	if d := r.URL.Query().Get("depth"); d != "" {
		var err error
		depth, err = strconv.Atoi(d)
		if err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "depth must be a valid integer", err))
			return
		}
	}

	// Parse optional evaluatorId — malformed UUID → 400, absent → nil
	var evaluatorID *uuid.UUID
	if eid := r.URL.Query().Get("evaluatorId"); eid != "" {
		parsed, err := uuid.Parse(eid)
		if err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "evaluatorId must be a valid UUID v4", err))
			return
		}
		evaluatorID = &parsed
	}

	// Parse optional headEmployeeId — find node headed by this employee
	var headEmployeeID *uuid.UUID
	if hid := r.URL.Query().Get("headEmployeeId"); hid != "" {
		parsed, err := uuid.Parse(hid)
		if err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "headEmployeeId must be a valid UUID v4", err))
			return
		}
		headEmployeeID = &parsed
	}

	result, err := h.treeSvc.GetTreeNodes(r.Context(), treeID, format, depth, evaluatorID, headEmployeeID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ExportOrgTree handles GET /api/v1/org-trees/{treeId}/export
func (h *OrgHandler) ExportOrgTree(w http.ResponseWriter, r *http.Request) {
	treeID := chi.URLParam(r, "treeId")
	if treeID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "treeId path parameter is required", nil))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")

	if err := h.treeSvc.ExportTree(r.Context(), treeID, w); err != nil {
		// Too late to change status code, but we can still log
		log.Printf("handler/org: export tree error: %v", err)
	}
}

// ==================== Org Node Endpoints ====================

// GetOrgNode handles GET /api/v1/org-nodes/{nodeId}
func (h *OrgHandler) GetOrgNode(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeId")
	if nodeID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "nodeId path parameter is required", nil))
		return
	}

	if _, err := uuid.Parse(nodeID); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "nodeId must be a valid UUID v4", err))
		return
	}

	result, err := h.nodeSvc.GetNode(r.Context(), nodeID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateOrgNode handles POST /api/v1/org-nodes
func (h *OrgHandler) CreateOrgNode(w http.ResponseWriter, r *http.Request) {
	var req org.CreateOrgNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "invalid JSON body", err))
		return
	}

	result, err := h.nodeSvc.CreateNode(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// UpdateOrgNode handles PUT /api/v1/org-nodes/{nodeId}
func (h *OrgHandler) UpdateOrgNode(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeId")
	if nodeID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "nodeId path parameter is required", nil))
		return
	}

	if _, err := uuid.Parse(nodeID); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "nodeId must be a valid UUID v4", err))
		return
	}

	var req org.UpdateOrgNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "invalid JSON body", err))
		return
	}

	result, err := h.nodeSvc.UpdateNode(r.Context(), nodeID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteOrgNode handles DELETE /api/v1/org-nodes/{nodeId}
func (h *OrgHandler) DeleteOrgNode(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeId")
	if nodeID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "nodeId path parameter is required", nil))
		return
	}

	if _, err := uuid.Parse(nodeID); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "nodeId must be a valid UUID v4", err))
		return
	}

	if err := h.nodeSvc.DeleteNode(r.Context(), nodeID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MoveOrgNode handles POST /api/v1/org-nodes/{nodeId}/move
func (h *OrgHandler) MoveOrgNode(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeId")
	if nodeID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "nodeId path parameter is required", nil))
		return
	}

	var req org.MoveOrgNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "invalid JSON body", err))
		return
	}

	if req.NewParentID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "newParentId is required", nil))
		return
	}

	result, err := h.nodeSvc.MoveNode(r.Context(), nodeID, req.NewParentID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ==================== Employee Endpoints ====================

// ListEmployees handles GET /api/v1/employees
func (h *OrgHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	treeID := q.Get("treeId")
	nodeID := q.Get("nodeId")
	profileID := q.Get("profileId")
	isActive := q.Get("isActive")
	query := q.Get("q")

	limit := 50
	if l := q.Get("limit"); l != "" {
		var err error
		limit, err = strconv.Atoi(l)
		if err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "limit must be a valid integer", err))
			return
		}
		if limit < 1 || limit > 200 {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "limit must be between 1 and 200", nil))
			return
		}
	}

	offset := 0
	if o := q.Get("offset"); o != "" {
		var err error
		offset, err = strconv.Atoi(o)
		if err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "offset must be a valid integer", err))
			return
		}
		if offset < 0 {
			offset = 0
		}
	}

	result, err := h.employeeSvc.ListEmployees(r.Context(), treeID, nodeID, profileID, isActive, query, offset, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetEmployee handles GET /api/v1/employees/{empId}
func (h *OrgHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	empID := chi.URLParam(r, "empId")
	if empID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "empId path parameter is required", nil))
		return
	}

	if _, err := uuid.Parse(empID); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "empId must be a valid UUID v4", err))
		return
	}

	result, err := h.employeeSvc.GetEmployee(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetMyEvaluatees handles GET /api/v1/employees/{empId}/evaluatees
func (h *OrgHandler) GetMyEvaluatees(w http.ResponseWriter, r *http.Request) {
	empID := chi.URLParam(r, "empId")
	if empID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "empId path parameter is required", nil))
		return
	}

	if _, err := uuid.Parse(empID); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "empId must be a valid UUID v4", err))
		return
	}

	q := r.URL.Query()

	limit := 50
	if l := q.Get("limit"); l != "" {
		var err error
		limit, err = strconv.Atoi(l)
		if err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "limit must be a valid integer", err))
			return
		}
		if limit < 1 || limit > 200 {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "limit must be between 1 and 200", nil))
			return
		}
	}

	offset := 0
	if o := q.Get("offset"); o != "" {
		var err error
		offset, err = strconv.Atoi(o)
		if err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "offset must be a valid integer", err))
			return
		}
		if offset < 0 {
			offset = 0
		}
	}

	searchQuery := q.Get("q")

	cycleID := q.Get("cycleId")
	if cycleID != "" {
		if _, err := uuid.Parse(strings.TrimSpace(cycleID)); err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "cycleId must be a valid UUID v4", err))
			return
		}
		cycleID = strings.TrimSpace(cycleID)
	}

	result, err := h.evaluateeSvc.GetMyEvaluateesPaginated(r.Context(), empID, searchQuery, offset, limit, cycleID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetTeam handles GET /api/v1/employees/{empId}/team
// Returns all employees in org nodes where the given employee is head_employee.
func (h *OrgHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	empID := chi.URLParam(r, "empId")
	if empID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "empId path parameter is required", nil))
		return
	}

	result, err := h.evaluateeSvc.GetTeamMembers(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetManager handles GET /api/v1/employees/{empId}/manager
func (h *OrgHandler) GetManager(w http.ResponseWriter, r *http.Request) {
	empID := chi.URLParam(r, "empId")
	if empID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "empId path parameter is required", nil))
		return
	}

	result, err := h.evaluateeSvc.GetManager(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetAncestors handles GET /api/v1/employees/{empId}/ancestors
func (h *OrgHandler) GetAncestors(w http.ResponseWriter, r *http.Request) {
	empID := chi.URLParam(r, "empId")
	if empID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "empId path parameter is required", nil))
		return
	}

	result, err := h.evaluateeSvc.GetChainOfCommand(r.Context(), empID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// BatchLookupEmployees handles POST /api/v1/employees/batch
func (h *OrgHandler) BatchLookupEmployees(w http.ResponseWriter, r *http.Request) {
	var req org.BatchEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "invalid JSON body", err))
		return
	}

	if len(req.IDs) == 0 {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "ids array is required", nil))
		return
	}

	if len(req.IDs) > 100 {
		req.IDs = req.IDs[:100]
	}

	result, err := h.evaluateeSvc.BatchLookup(r.Context(), req.IDs)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// SearchEmployees handles GET /api/v1/employees/search
func (h *OrgHandler) SearchEmployees(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" || len(strings.TrimSpace(q)) < 2 {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "search query (q) must be at least 2 characters", nil))
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		var err error
		limit, err = strconv.Atoi(l)
		if err != nil {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "limit must be a valid integer", err))
			return
		}
		if limit < 1 || limit > 50 {
			writeError(w, errors.NewDomainError(errors.InvalidRequest, "limit must be between 1 and 50", nil))
			return
		}
	}

	result, err := h.employeeSvc.SearchEmployees(r.Context(), q, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateEmployee handles PUT /api/v1/employees/{empId}
func (h *OrgHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	empID := chi.URLParam(r, "empId")
	if empID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "empId path parameter is required", nil))
		return
	}

	if _, err := uuid.Parse(empID); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "empId must be a valid UUID v4", err))
		return
	}

	var req org.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "invalid JSON body", err))
		return
	}

	if req.ProfileID == "" || req.OrgNodeID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "profileId and orgNodeId are required", nil))
		return
	}

	updatedBy, ok := auth.GetEmployeeID(r.Context())
	if !ok {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "unable to identify current user", nil))
		return
	}

	result, err := h.employeeSvc.UpdateEmployee(r.Context(), empID, req, updatedBy)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ==================== Area Metrics Endpoints ====================

// GetAreaMetrics handles GET /api/v1/org-nodes/{nodeId}/area-metrics
func (h *OrgHandler) GetAreaMetrics(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeId")
	if nodeID == "" {
		writeError(w, errors.NewDomainError(errors.InvalidRequest, "nodeId path parameter is required", nil))
		return
	}

	cycleID := r.URL.Query().Get("cycleId")

	result, err := h.metricsSvc.GetAreaMetrics(r.Context(), nodeID, cycleID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
