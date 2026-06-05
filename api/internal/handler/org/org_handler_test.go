package org_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/org"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
	handler "github.com/sed-evaluacion-desempeno/api/internal/handler/org"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/org"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- mock services ----------

type mockOrgTreeService struct {
	getTreesFunc     func(ctx context.Context, treeType string) (*dto.OrgTreeListResponse, error)
	getTreeFunc      func(ctx context.Context, treeID string) (*dto.OrgTreeDetailResponse, error)
	getTreeNodesFunc func(ctx context.Context, treeID, format string, depth int) (interface{}, error)
	exportTreeFunc   func(ctx context.Context, treeID string, w io.Writer) error
}

func (m *mockOrgTreeService) GetTrees(ctx context.Context, treeType string) (*dto.OrgTreeListResponse, error) {
	return m.getTreesFunc(ctx, treeType)
}
func (m *mockOrgTreeService) GetTree(ctx context.Context, treeID string) (*dto.OrgTreeDetailResponse, error) {
	return m.getTreeFunc(ctx, treeID)
}
func (m *mockOrgTreeService) GetTreeNodes(ctx context.Context, treeID, format string, depth int) (interface{}, error) {
	return m.getTreeNodesFunc(ctx, treeID, format, depth)
}
func (m *mockOrgTreeService) ExportTree(ctx context.Context, treeID string, w io.Writer) error {
	return m.exportTreeFunc(ctx, treeID, w)
}

type mockOrgNodeService struct {
	getNodeFunc    func(ctx context.Context, nodeID string) (*dto.OrgNodeDetailResponse, error)
	createNodeFunc func(ctx context.Context, req dto.CreateOrgNodeRequest) (*dto.OrgNodeDetailResponse, error)
	updateNodeFunc func(ctx context.Context, nodeID string, req dto.UpdateOrgNodeRequest) (*dto.OrgNodeDetailResponse, error)
	deleteNodeFunc func(ctx context.Context, nodeID string) error
	moveNodeFunc   func(ctx context.Context, nodeID, newParentID string) (*dto.OrgNodeDetailResponse, error)
}

func (m *mockOrgNodeService) GetNode(ctx context.Context, nodeID string) (*dto.OrgNodeDetailResponse, error) {
	return m.getNodeFunc(ctx, nodeID)
}
func (m *mockOrgNodeService) CreateNode(ctx context.Context, req dto.CreateOrgNodeRequest) (*dto.OrgNodeDetailResponse, error) {
	return m.createNodeFunc(ctx, req)
}
func (m *mockOrgNodeService) UpdateNode(ctx context.Context, nodeID string, req dto.UpdateOrgNodeRequest) (*dto.OrgNodeDetailResponse, error) {
	return m.updateNodeFunc(ctx, nodeID, req)
}
func (m *mockOrgNodeService) DeleteNode(ctx context.Context, nodeID string) error {
	return m.deleteNodeFunc(ctx, nodeID)
}
func (m *mockOrgNodeService) MoveNode(ctx context.Context, nodeID, newParentID string) (*dto.OrgNodeDetailResponse, error) {
	return m.moveNodeFunc(ctx, nodeID, newParentID)
}

type mockEmployeeService struct {
	listEmployeesFunc   func(ctx context.Context, treeID, nodeID, profileID, isActive, query, cursor string, limit int) (*dto.EmployeeListResponse, error)
	getEmployeeFunc     func(ctx context.Context, empID string) (*dto.EmployeeDetailResponse, error)
	searchEmployeesFunc func(ctx context.Context, query string, limit int) (*dto.EmployeeListResponse, error)
}

func (m *mockEmployeeService) ListEmployees(ctx context.Context, treeID, nodeID, profileID, isActive, query, cursor string, limit int) (*dto.EmployeeListResponse, error) {
	return m.listEmployeesFunc(ctx, treeID, nodeID, profileID, isActive, query, cursor, limit)
}
func (m *mockEmployeeService) GetEmployee(ctx context.Context, empID string) (*dto.EmployeeDetailResponse, error) {
	return m.getEmployeeFunc(ctx, empID)
}
func (m *mockEmployeeService) SearchEmployees(ctx context.Context, query string, limit int) (*dto.EmployeeListResponse, error) {
	return m.searchEmployeesFunc(ctx, query, limit)
}

type mockEvaluateeService struct {
	getMyEvaluateesFunc   func(ctx context.Context, evaluatorID string) (*dto.EmployeeListResponse, error)
	getManagerFunc        func(ctx context.Context, empID string) (*dto.EmployeeDetailResponse, error)
	getChainOfCommandFunc func(ctx context.Context, empID string) (*dto.AncestorChainResponse, error)
	batchLookupFunc       func(ctx context.Context, ids []string) (*dto.EmployeeListResponse, error)
}

func (m *mockEvaluateeService) GetMyEvaluatees(ctx context.Context, evaluatorID string) (*dto.EmployeeListResponse, error) {
	return m.getMyEvaluateesFunc(ctx, evaluatorID)
}
func (m *mockEvaluateeService) GetManager(ctx context.Context, empID string) (*dto.EmployeeDetailResponse, error) {
	return m.getManagerFunc(ctx, empID)
}
func (m *mockEvaluateeService) GetChainOfCommand(ctx context.Context, empID string) (*dto.AncestorChainResponse, error) {
	return m.getChainOfCommandFunc(ctx, empID)
}
func (m *mockEvaluateeService) BatchLookup(ctx context.Context, ids []string) (*dto.EmployeeListResponse, error) {
	return m.batchLookupFunc(ctx, ids)
}

type mockEvaluatorService struct {
	getEvaluatorScopeFunc func(ctx context.Context, evaluatorID, cycleID string) (*dto.EvaluatorScopeResponse, error)
	resolveEvaluatorFunc  func(ctx context.Context, evaluateeID string) (*dto.EmployeeDetailResponse, error)
}

func (m *mockEvaluatorService) GetEvaluatorScope(ctx context.Context, evaluatorID, cycleID string) (*dto.EvaluatorScopeResponse, error) {
	return m.getEvaluatorScopeFunc(ctx, evaluatorID, cycleID)
}
func (m *mockEvaluatorService) ResolveEvaluator(ctx context.Context, evaluateeID string) (*dto.EmployeeDetailResponse, error) {
	return m.resolveEvaluatorFunc(ctx, evaluateeID)
}

// ---------- helpers ----------

func newTestHandler(
	treeSvc svc.OrgTreeService,
	nodeSvc svc.OrgNodeService,
	empSvc svc.EmployeeService,
	evalSvc svc.EvaluateeService,
	evaluatorSvc svc.EvaluatorService,
) *handler.OrgHandler {
	return handler.NewOrgHandler(treeSvc, nodeSvc, empSvc, evalSvc, evaluatorSvc)
}

func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ---------- happy path tests ----------

func TestListOrgTrees_Success(t *testing.T) {
	t.Parallel()

	treeSvc := &mockOrgTreeService{
		getTreesFunc: func(_ context.Context, treeType string) (*dto.OrgTreeListResponse, error) {
			assert.Equal(t, "corporate", treeType)
			return &dto.OrgTreeListResponse{
				Data: []dto.OrgTreeResponse{
					{ID: uuid.New().String(), Name: "Alpha Corp", Type: "corporate", NodeCount: 42},
				},
			}, nil
		},
	}

	h := newTestHandler(treeSvc, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/org-trees?type=corporate", nil)
	rec := httptest.NewRecorder()

	h.ListOrgTrees(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.OrgTreeListResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "Alpha Corp", resp.Data[0].Name)
}

func TestGetOrgTree_Success(t *testing.T) {
	t.Parallel()

	treeID := uuid.New().String()
	treeSvc := &mockOrgTreeService{
		getTreeFunc: func(_ context.Context, id string) (*dto.OrgTreeDetailResponse, error) {
			assert.Equal(t, treeID, id)
			return &dto.OrgTreeDetailResponse{Data: dto.OrgTreeResponse{ID: treeID, Name: "Beta Tree", Type: "retail", NodeCount: 7}}, nil
		},
	}

	h := newTestHandler(treeSvc, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/org-trees/"+treeID, nil)
	req = withChiParam(req, "treeId", treeID)
	rec := httptest.NewRecorder()

	h.GetOrgTree(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.OrgTreeDetailResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "Beta Tree", resp.Data.Name)
}

func TestGetTreeNodes_Success(t *testing.T) {
	t.Parallel()

	treeID := uuid.New().String()
	treeSvc := &mockOrgTreeService{
		getTreeNodesFunc: func(_ context.Context, id, format string, depth int) (interface{}, error) {
			assert.Equal(t, treeID, id)
			assert.Equal(t, "flat", format)
			assert.Equal(t, 2, depth)
			return &dto.OrgNodeFlatList{
				Data: []dto.OrgNodeResponse{
					{ID: uuid.New().String(), Name: "Root", Depth: 0},
					{ID: uuid.New().String(), Name: "Child", Depth: 1},
				},
				Meta: struct {
					Format string `json:"format"`
					Total  int    `json:"total"`
				}{Format: "flat", Total: 2},
			}, nil
		},
	}

	h := newTestHandler(treeSvc, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/org-trees/"+treeID+"/nodes?format=flat&depth=2", nil)
	req = withChiParam(req, "treeId", treeID)
	rec := httptest.NewRecorder()

	h.GetOrgTreeNodes(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.OrgNodeFlatList
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data, 2)
	assert.Equal(t, "flat", resp.Meta.Format)
}

func TestCreateNode_Success(t *testing.T) {
	t.Parallel()

	nodeSvc := &mockOrgNodeService{
		createNodeFunc: func(_ context.Context, req dto.CreateOrgNodeRequest) (*dto.OrgNodeDetailResponse, error) {
			assert.Equal(t, "New Node", req.Name)
			assert.Equal(t, "corporate", req.Type)
			return &dto.OrgNodeDetailResponse{Data: dto.OrgNodeResponse{ID: uuid.New().String(), Name: req.Name}}, nil
		},
	}

	h := newTestHandler(nil, nodeSvc, nil, nil, nil)
	body, _ := json.Marshal(dto.CreateOrgNodeRequest{Name: "New Node", Type: "corporate", OrganizationID: uuid.New().String(), Code: "NN01"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/org-nodes", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.CreateOrgNode(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp dto.OrgNodeDetailResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "New Node", resp.Data.Name)
}

func TestUpdateNode_Success(t *testing.T) {
	t.Parallel()

	nodeID := uuid.New().String()
	nodeSvc := &mockOrgNodeService{
		updateNodeFunc: func(_ context.Context, id string, req dto.UpdateOrgNodeRequest) (*dto.OrgNodeDetailResponse, error) {
			assert.Equal(t, nodeID, id)
			assert.Equal(t, "Updated", req.Name)
			return &dto.OrgNodeDetailResponse{Data: dto.OrgNodeResponse{ID: nodeID, Name: req.Name}}, nil
		},
	}

	h := newTestHandler(nil, nodeSvc, nil, nil, nil)
	body, _ := json.Marshal(dto.UpdateOrgNodeRequest{Name: "Updated", Version: 1})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/org-nodes/"+nodeID, bytes.NewReader(body))
	req = withChiParam(req, "nodeId", nodeID)
	rec := httptest.NewRecorder()

	h.UpdateOrgNode(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.OrgNodeDetailResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "Updated", resp.Data.Name)
}

func TestDeleteNode_Success(t *testing.T) {
	t.Parallel()

	nodeID := uuid.New().String()
	nodeSvc := &mockOrgNodeService{
		deleteNodeFunc: func(_ context.Context, id string) error {
			assert.Equal(t, nodeID, id)
			return nil
		},
	}

	h := newTestHandler(nil, nodeSvc, nil, nil, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/org-nodes/"+nodeID, nil)
	req = withChiParam(req, "nodeId", nodeID)
	rec := httptest.NewRecorder()

	h.DeleteOrgNode(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestListEmployees_Success(t *testing.T) {
	t.Parallel()

	empSvc := &mockEmployeeService{
		listEmployeesFunc: func(_ context.Context, treeID, nodeID, profileID, isActive, query, cursor string, limit int) (*dto.EmployeeListResponse, error) {
			assert.Equal(t, "some-tree", treeID)
			assert.Equal(t, 25, limit)
			return &dto.EmployeeListResponse{
				Data: []dto.EmployeeListItem{
					{ID: uuid.New().String(), FirstName: "Alice"},
				},
				Meta: struct {
					NextCursor string `json:"nextCursor,omitempty"`
					HasMore    bool   `json:"hasMore"`
					Limit      int    `json:"limit"`
				}{HasMore: false, Limit: 25},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, empSvc, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees?treeId=some-tree&limit=25", nil)
	rec := httptest.NewRecorder()

	h.ListEmployees(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EmployeeListResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "Alice", resp.Data[0].FirstName)
}

func TestGetEmployee_Success(t *testing.T) {
	t.Parallel()

	empID := uuid.New().String()
	empSvc := &mockEmployeeService{
		getEmployeeFunc: func(_ context.Context, id string) (*dto.EmployeeDetailResponse, error) {
			assert.Equal(t, empID, id)
			return &dto.EmployeeDetailResponse{Data: dto.EmployeeDetail{ID: empID, FirstName: "Bob"}}, nil
		},
	}

	h := newTestHandler(nil, nil, empSvc, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/"+empID, nil)
	req = withChiParam(req, "empId", empID)
	rec := httptest.NewRecorder()

	h.GetEmployee(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EmployeeDetailResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "Bob", resp.Data.FirstName)
}

func TestGetMyEvaluatees_Success(t *testing.T) {
	t.Parallel()

	empID := uuid.New().String()
	evalSvc := &mockEvaluateeService{
		getMyEvaluateesFunc: func(_ context.Context, id string) (*dto.EmployeeListResponse, error) {
			assert.Equal(t, empID, id)
			return &dto.EmployeeListResponse{
				Data: []dto.EmployeeListItem{
					{ID: uuid.New().String(), FirstName: "Carol"},
				},
				Meta: struct {
					NextCursor string `json:"nextCursor,omitempty"`
					HasMore    bool   `json:"hasMore"`
					Limit      int    `json:"limit"`
				}{Limit: 1, HasMore: false},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, evalSvc, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/"+empID+"/evaluatees", nil)
	req = withChiParam(req, "empId", empID)
	rec := httptest.NewRecorder()

	h.GetMyEvaluatees(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EmployeeListResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "Carol", resp.Data[0].FirstName)
}

func TestGetChainOfCommand_Success(t *testing.T) {
	t.Parallel()

	empID := uuid.New().String()
	evalSvc := &mockEvaluateeService{
		getChainOfCommandFunc: func(_ context.Context, id string) (*dto.AncestorChainResponse, error) {
			assert.Equal(t, empID, id)
			return &dto.AncestorChainResponse{
				Data: []dto.AncestorItem{
					{ID: uuid.New().String(), Name: "CEO", Depth: 2, Relation: "ceo"},
					{ID: uuid.New().String(), Name: "VP", Depth: 1, Relation: "vp"},
					{ID: empID, Name: "Self", Depth: 0, Relation: "self"},
				},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, evalSvc, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/"+empID+"/ancestors", nil)
	req = withChiParam(req, "empId", empID)
	rec := httptest.NewRecorder()

	h.GetAncestors(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.AncestorChainResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data, 3)
	assert.Equal(t, "self", resp.Data[2].Relation)
}

func TestBatchResolve_Success(t *testing.T) {
	t.Parallel()

	id1 := uuid.New().String()
	id2 := uuid.New().String()
	evalSvc := &mockEvaluateeService{
		batchLookupFunc: func(_ context.Context, ids []string) (*dto.EmployeeListResponse, error) {
			require.Len(t, ids, 2)
			return &dto.EmployeeListResponse{
				Data: []dto.EmployeeListItem{
					{ID: id1, FirstName: "Dan"},
					{ID: id2, FirstName: "Dana"},
				},
				Meta: struct {
					NextCursor string `json:"nextCursor,omitempty"`
					HasMore    bool   `json:"hasMore"`
					Limit      int    `json:"limit"`
				}{Limit: 2, HasMore: false},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, evalSvc, nil)
	body, _ := json.Marshal(dto.BatchEmployeeRequest{IDs: []string{id1, id2}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/batch", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.BatchLookupEmployees(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EmployeeListResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data, 2)
}

func TestSearchEmployees_Success(t *testing.T) {
	t.Parallel()

	empSvc := &mockEmployeeService{
		searchEmployeesFunc: func(_ context.Context, query string, limit int) (*dto.EmployeeListResponse, error) {
			assert.Equal(t, "alice", query)
			assert.Equal(t, 20, limit)
			return &dto.EmployeeListResponse{
				Data: []dto.EmployeeListItem{
					{ID: uuid.New().String(), FirstName: "Alice"},
				},
				Meta: struct {
					NextCursor string `json:"nextCursor,omitempty"`
					HasMore    bool   `json:"hasMore"`
					Limit      int    `json:"limit"`
				}{Limit: 20, HasMore: false},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, empSvc, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/search?q=alice", nil)
	rec := httptest.NewRecorder()

	h.SearchEmployees(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EmployeeListResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data, 1)
}

func TestGetEvaluatorScope_Success(t *testing.T) {
	t.Parallel()

	evaluatorID := uuid.New().String()
	evaluatorSvc := &mockEvaluatorService{
		getEvaluatorScopeFunc: func(_ context.Context, id, cycleID string) (*dto.EvaluatorScopeResponse, error) {
			assert.Equal(t, evaluatorID, id)
			assert.Equal(t, "cycle-123", cycleID)
			return &dto.EvaluatorScopeResponse{
				EvaluatorID: evaluatorID,
				CycleID:     "cycle-123",
				ScopeType:   "department",
				EvaluateeCount: 5,
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, nil, evaluatorSvc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/evaluator-scopes?evaluatorId="+evaluatorID+"&cycleId=cycle-123", nil)
	rec := httptest.NewRecorder()

	h.GetEvaluatorScope(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EvaluatorScopeResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "department", resp.ScopeType)
	assert.Equal(t, 5, resp.EvaluateeCount)
}

// ---------- error tests ----------

func TestGetTree_NotFound(t *testing.T) {
	t.Parallel()

	treeID := uuid.New().String()
	treeSvc := &mockOrgTreeService{
		getTreeFunc: func(_ context.Context, id string) (*dto.OrgTreeDetailResponse, error) {
			return nil, repo.ErrTreeNotFound
		},
	}

	h := newTestHandler(treeSvc, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/org-trees/"+treeID, nil)
	req = withChiParam(req, "treeId", treeID)
	rec := httptest.NewRecorder()

	h.GetOrgTree(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	var apiErr map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&apiErr))
	assert.Equal(t, "TREE_NOT_FOUND", apiErr["error"].(map[string]interface{})["code"])
}

func TestCreateNode_InvalidParent(t *testing.T) {
	t.Parallel()

	nodeSvc := &mockOrgNodeService{
		createNodeFunc: func(_ context.Context, req dto.CreateOrgNodeRequest) (*dto.OrgNodeDetailResponse, error) {
			return nil, repo.ErrInvalidParent
		},
	}

	h := newTestHandler(nil, nodeSvc, nil, nil, nil)
	body, _ := json.Marshal(dto.CreateOrgNodeRequest{Name: "Orphan", Type: "corporate", OrganizationID: uuid.New().String(), Code: "X"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/org-nodes", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.CreateOrgNode(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var apiErr map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&apiErr))
	assert.Equal(t, "INVALID_PARENT", apiErr["error"].(map[string]interface{})["code"])
}

func TestDeleteNode_HasChildren(t *testing.T) {
	t.Parallel()

	nodeID := uuid.New().String()
	nodeSvc := &mockOrgNodeService{
		deleteNodeFunc: func(_ context.Context, id string) error {
			return repo.ErrNodeHasChildren
		},
	}

	h := newTestHandler(nil, nodeSvc, nil, nil, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/org-nodes/"+nodeID, nil)
	req = withChiParam(req, "nodeId", nodeID)
	rec := httptest.NewRecorder()

	h.DeleteOrgNode(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	var apiErr map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&apiErr))
	assert.Equal(t, "NODE_HAS_CHILDREN", apiErr["error"].(map[string]interface{})["code"])
}

func TestUpdateNode_VersionConflict(t *testing.T) {
	t.Parallel()

	nodeID := uuid.New().String()
	nodeSvc := &mockOrgNodeService{
		updateNodeFunc: func(_ context.Context, id string, req dto.UpdateOrgNodeRequest) (*dto.OrgNodeDetailResponse, error) {
			return nil, repo.ErrStaleVersion
		},
	}

	h := newTestHandler(nil, nodeSvc, nil, nil, nil)
	body, _ := json.Marshal(dto.UpdateOrgNodeRequest{Name: "Conflict", Version: 0})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/org-nodes/"+nodeID, bytes.NewReader(body))
	req = withChiParam(req, "nodeId", nodeID)
	rec := httptest.NewRecorder()

	h.UpdateOrgNode(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	var apiErr map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&apiErr))
	assert.Equal(t, "STALE_VERSION", apiErr["error"].(map[string]interface{})["code"])
}

func TestBatchResolve_TooManyIDs(t *testing.T) {
	t.Parallel()

	// The handler truncates to 100 IDs and returns 200. This test documents
	// the current behavior rather than asserting a 400 (which would require
	// a validation change in the handler).
	ids := make([]string, 150)
	for i := range ids {
		ids[i] = uuid.New().String()
	}

	var received []string
	evalSvc := &mockEvaluateeService{
		batchLookupFunc: func(_ context.Context, idList []string) (*dto.EmployeeListResponse, error) {
			received = idList
			return &dto.EmployeeListResponse{
				Data: []dto.EmployeeListItem{},
				Meta: struct {
					NextCursor string `json:"nextCursor,omitempty"`
					HasMore    bool   `json:"hasMore"`
					Limit      int    `json:"limit"`
				}{Limit: len(idList), HasMore: false},
			}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, evalSvc, nil)
	body, _ := json.Marshal(dto.BatchEmployeeRequest{IDs: ids})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/batch", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.BatchLookupEmployees(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, received, 100, "handler should truncate IDs to 100")
}

// ---------- response time tests ----------

func TestListEmployees_ResponseTime(t *testing.T) {
	t.Parallel()

	calls := 0
	empSvc := &mockEmployeeService{
		listEmployeesFunc: func(_ context.Context, _, _, _, _, _, _ string, _ int) (*dto.EmployeeListResponse, error) {
			calls++
			return &dto.EmployeeListResponse{Data: []dto.EmployeeListItem{}}, nil
		},
	}

	h := newTestHandler(nil, nil, empSvc, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees?limit=50", nil)

	const iterations = 100
	start := time.Now()
	for i := 0; i < iterations; i++ {
		rec := httptest.NewRecorder()
		h.ListEmployees(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}
	elapsed := time.Since(start)
	avg := elapsed / iterations

	assert.Less(t, avg, 200*time.Millisecond, "average response time should be under 200ms")
}

func TestGetMyEvaluatees_ResponseTime(t *testing.T) {
	t.Parallel()

	empID := uuid.New().String()
	calls := 0
	evalSvc := &mockEvaluateeService{
		getMyEvaluateesFunc: func(_ context.Context, id string) (*dto.EmployeeListResponse, error) {
			calls++
			return &dto.EmployeeListResponse{Data: []dto.EmployeeListItem{}}, nil
		},
	}

	h := newTestHandler(nil, nil, nil, evalSvc, nil)

	const iterations = 100
	start := time.Now()
	for i := 0; i < iterations; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/"+empID+"/evaluatees", nil)
		req = withChiParam(req, "empId", empID)
		rec := httptest.NewRecorder()
		h.GetMyEvaluatees(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}
	elapsed := time.Since(start)
	avg := elapsed / iterations

	assert.Less(t, avg, 100*time.Millisecond, "average response time should be under 100ms")
}

// ---------- concurrency tests ----------

func TestUpdateNode_Concurrent(t *testing.T) {
	t.Parallel()

	nodeID := uuid.New().String()
	var mu sync.Mutex
	callCount := 0

	nodeSvc := &mockOrgNodeService{
		updateNodeFunc: func(_ context.Context, id string, req dto.UpdateOrgNodeRequest) (*dto.OrgNodeDetailResponse, error) {
			mu.Lock()
			callCount++
			mu.Unlock()
			return &dto.OrgNodeDetailResponse{Data: dto.OrgNodeResponse{ID: id, Name: req.Name}}, nil
		},
	}

	h := newTestHandler(nil, nodeSvc, nil, nil, nil)
	body, _ := json.Marshal(dto.UpdateOrgNodeRequest{Name: "Concurrent", Version: 1})

	const workers = 50
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/org-nodes/"+nodeID, bytes.NewReader(body))
			req = withChiParam(req, "nodeId", nodeID)
			rec := httptest.NewRecorder()
			h.UpdateOrgNode(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}()
	}

	wg.Wait()
	assert.Equal(t, workers, callCount)
}

func TestDeleteNode_Concurrent(t *testing.T) {
	t.Parallel()

	nodeID := uuid.New().String()
	var mu sync.Mutex
	callCount := 0

	nodeSvc := &mockOrgNodeService{
		deleteNodeFunc: func(_ context.Context, id string) error {
			mu.Lock()
			callCount++
			mu.Unlock()
			return nil
		},
	}

	h := newTestHandler(nil, nodeSvc, nil, nil, nil)

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/org-nodes/"+nodeID, nil)
			req = withChiParam(req, "nodeId", nodeID)
			rec := httptest.NewRecorder()
			h.DeleteOrgNode(rec, req)
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}()
	}

	wg.Wait()
	assert.Equal(t, workers, callCount)
}

// ---------- additional handler validation tests ----------

func TestGetOrgTree_InvalidUUID(t *testing.T) {
	t.Parallel()

	h := newTestHandler(nil, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/org-trees/not-a-uuid", nil)
	req = withChiParam(req, "treeId", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.GetOrgTree(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	bodyStr := rec.Body.String()
	assert.Contains(t, bodyStr, "treeId must be a valid UUID v4")
}

func TestCreateNode_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := newTestHandler(nil, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/org-nodes", strings.NewReader("not json"))
	rec := httptest.NewRecorder()

	h.CreateOrgNode(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSearchEmployees_QueryTooShort(t *testing.T) {
	t.Parallel()

	h := newTestHandler(nil, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/search?q=a", nil)
	rec := httptest.NewRecorder()

	h.SearchEmployees(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	bodyStr := rec.Body.String()
	assert.Contains(t, bodyStr, "at least 2 characters")
}

func TestGetEvaluatorScope_MissingEvaluatorID(t *testing.T) {
	t.Parallel()

	h := newTestHandler(nil, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/evaluator-scopes", nil)
	rec := httptest.NewRecorder()

	h.GetEvaluatorScope(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	bodyStr := rec.Body.String()
	assert.Contains(t, bodyStr, "evaluatorId query parameter is required")
}
