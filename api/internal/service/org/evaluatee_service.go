package org

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/dto/org"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
)

// EvaluateeService defines the interface for evaluatee and chain-of-command operations.
type EvaluateeService interface {
	GetMyEvaluatees(ctx context.Context, evaluatorID string) (*org.EmployeeListResponse, error)
	GetManager(ctx context.Context, empID string) (*org.EmployeeDetailResponse, error)
	GetChainOfCommand(ctx context.Context, empID string) (*org.AncestorChainResponse, error)
	BatchLookup(ctx context.Context, ids []string) (*org.EmployeeListResponse, error)
}

// EvaluatorService defines the interface for evaluator scope operations.
type EvaluatorService interface {
	GetEvaluatorScope(ctx context.Context, evaluatorID, cycleID string) (*org.EvaluatorScopeResponse, error)
	ResolveEvaluator(ctx context.Context, evaluateeID string) (*org.EmployeeDetailResponse, error)
}

type evaluateeService struct {
	empRepo   *repo.EmployeeRepo
	nodeRepo  *repo.OrgNodeRepo
	scopeRepo *repo.EvaluatorScopeRepo
	client    *internal.Client
}

// NewEvaluateeService creates a new EvaluateeService.
func NewEvaluateeService(empRepo *repo.EmployeeRepo, nodeRepo *repo.OrgNodeRepo, scopeRepo *repo.EvaluatorScopeRepo, client *internal.Client) EvaluateeService {
	return &evaluateeService{
		empRepo:   empRepo,
		nodeRepo:  nodeRepo,
		scopeRepo: scopeRepo,
		client:    client,
	}
}

// NewEvaluatorService creates a new EvaluatorService (uses same underlying service).
func NewEvaluatorService(empRepo *repo.EmployeeRepo, nodeRepo *repo.OrgNodeRepo, scopeRepo *repo.EvaluatorScopeRepo, client *internal.Client) EvaluatorService {
	return &evaluateeService{
		empRepo:   empRepo,
		nodeRepo:  nodeRepo,
		scopeRepo: scopeRepo,
		client:    client,
	}
}

func (s *evaluateeService) GetMyEvaluatees(ctx context.Context, evaluatorID string) (*org.EmployeeListResponse, error) {
	id, err := uuid.Parse(evaluatorID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid evaluator ID: must be a valid UUID", err)
	}

	// Verify evaluator exists
	_, err = s.empRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Direct reports only, active
	rows, err := s.empRepo.ListByManager(ctx, id, true)
	if err != nil {
		return nil, err
	}

	resp := &org.EmployeeListResponse{
		Data: make([]org.EmployeeListItem, len(rows)),
	}
	resp.Meta.Limit = len(rows)
	resp.Meta.HasMore = false

	for i, r := range rows {
		resp.Data[i] = employeeRowToItem(r)
	}

	return resp, nil
}

func (s *evaluateeService) GetManager(ctx context.Context, empID string) (*org.EmployeeDetailResponse, error) {
	id, err := uuid.Parse(empID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid employee ID: must be a valid UUID", err)
	}

	// Get the manager
	row, err := s.empRepo.GetManager(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		// Employee exists but has no manager
		return nil, repo.ErrEmployeeNotFound.WithDetails("Employee has no manager")
	}

	detail, err := s.empRepo.GetDetailByID(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	return buildEmployeeDetailResponse(detail), nil
}

func (s *evaluateeService) GetChainOfCommand(ctx context.Context, empID string) (*org.AncestorChainResponse, error) {
	id, err := uuid.Parse(empID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid employee ID: must be a valid UUID", err)
	}

	// Get employee to find their org node
	emp, err := s.empRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get node path
	node, err := s.nodeRepo.GetByID(ctx, emp.OrgNodeID)
	if err != nil {
		return nil, err
	}

	// Get ancestors via path
	ancestors, err := s.nodeRepo.GetAncestors(ctx, node.Path)
	if err != nil {
		return nil, err
	}

	// Build chain from root to self
	chain := &org.AncestorChainResponse{
		Data: make([]org.AncestorItem, 0, len(ancestors)),
	}

	totalDepth := len(ancestors)
	for i, a := range ancestors {
		depth := totalDepth - i - 1
		relation := resolveRelation(depth, totalDepth)

		chain.Data = append(chain.Data, org.AncestorItem{
			ID:       a.ID.String(),
			Name:     a.Name,
			Depth:    depth,
			Relation: relation,
		})
	}

	return chain, nil
}

func (s *evaluateeService) BatchLookup(ctx context.Context, ids []string) (*org.EmployeeListResponse, error) {
	if len(ids) == 0 {
		return &org.EmployeeListResponse{Data: []org.EmployeeListItem{}}, nil
	}

	uuidIDs := make([]uuid.UUID, 0, len(ids))
	for _, idStr := range ids {
		id, err := uuid.Parse(idStr)
		if err == nil {
			uuidIDs = append(uuidIDs, id)
		}
	}

	if len(uuidIDs) > 100 {
		uuidIDs = uuidIDs[:100]
	}

	rows, err := s.empRepo.GetByIDs(ctx, uuidIDs)
	if err != nil {
		return nil, err
	}

	// Maintain input order
	byID := make(map[string]*repo.EmployeeRow, len(rows))
	for _, r := range rows {
		byID[r.ID.String()] = r
	}

	resp := &org.EmployeeListResponse{
		Data: make([]org.EmployeeListItem, 0, len(uuidIDs)),
	}
	resp.Meta.Limit = len(uuidIDs)
	resp.Meta.HasMore = false

	for _, id := range uuidIDs {
		if r, ok := byID[id.String()]; ok {
			resp.Data = append(resp.Data, employeeRowToItem(r))
		}
	}

	return resp, nil
}

func (s *evaluateeService) GetEvaluatorScope(ctx context.Context, evaluatorID, cycleID string) (*org.EvaluatorScopeResponse, error) {
	evalID, err := uuid.Parse(evaluatorID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid evaluator ID: must be a valid UUID", err)
	}

	// Verify evaluator exists
	_, err = s.empRepo.GetByID(ctx, evalID)
	if err != nil {
		return nil, err
	}

	if cycleID == "" {
		// Find scopes without cycle filter
		rows, err := s.scopeRepo.GetByEvaluator(ctx, evalID)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, repo.ErrScopeNotFound
		}
		return scopeRowToResponse(rows[0]), nil
	}

	cycID, err := uuid.Parse(cycleID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid cycle ID: must be a valid UUID", err)
	}

	row, err := s.scopeRepo.GetByEvaluatorAndCycle(ctx, evalID, cycID)
	if err != nil {
		return nil, err
	}

	return scopeRowToResponse(row), nil
}

func (s *evaluateeService) ResolveEvaluator(ctx context.Context, evaluateeID string) (*org.EmployeeDetailResponse, error) {
	id, err := uuid.Parse(evaluateeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid evaluatee ID: must be a valid UUID", err)
	}

	// Get the evaluatee's manager (who evaluates them)
	manager, err := s.empRepo.GetManager(ctx, id)
	if err != nil {
		return nil, err
	}
	if manager == nil {
		return nil, repo.ErrEmployeeNotFound.WithDetails("Employee has no manager (root)")
	}

	detail, err := s.empRepo.GetDetailByID(ctx, manager.ID)
	if err != nil {
		return nil, err
	}

	return buildEmployeeDetailResponse(detail), nil
}

// ---------- helpers ----------

func resolveRelation(depth, totalDepth int) string {
	if depth == 0 {
		return "self"
	}
	if depth == totalDepth-1 {
		return "ceo"
	}
	if depth == totalDepth-2 {
		return "vp"
	}
	if depth == totalDepth-3 {
		return "director"
	}
	return "direct_manager"
}

func scopeRowToResponse(row *repo.EvaluatorScopeRow) *org.EvaluatorScopeResponse {
	resp := &org.EvaluatorScopeResponse{
		EvaluatorID: row.EvaluatorID.String(),
		ScopeType:   row.ScopeType,
	}

	if row.CycleID != nil {
		resp.CycleID = row.CycleID.String()
	}

	// Extract scopeData
	if data, ok := row.ScopeData["orgNodeIds"]; ok {
		if ids, ok := data.([]interface{}); ok {
			resp.ScopeData.OrgNodeIDs = make([]string, len(ids))
			for i, id := range ids {
				resp.ScopeData.OrgNodeIDs[i] = toString(id)
			}
		}
	}
	if data, ok := row.ScopeData["employeeIds"]; ok {
		if ids, ok := data.([]interface{}); ok {
			resp.ScopeData.EmployeeIDs = make([]string, len(ids))
			for i, id := range ids {
				resp.ScopeData.EmployeeIDs[i] = toString(id)
			}
		}
	}

	// Count evaluatees from scopeData
	if empIDs, ok := row.ScopeData["employeeIds"]; ok {
		if ids, ok := empIDs.([]interface{}); ok {
			resp.EvaluateeCount = len(ids)
		}
	}

	return resp
}

func buildEmployeeDetailResponse(detail *repo.EmployeeDetailRow) *org.EmployeeDetailResponse {
	resp := &org.EmployeeDetailResponse{
		Data: org.EmployeeDetail{
			ID:             detail.ID.String(),
			FirstName:      detail.FirstName,
			LastName:       detail.LastName,
			Email:          detail.Email,
			EmployeeNumber: detail.EmployeeNumber,
			OrgNodeID:      detail.OrgNodeID.String(),
			ProfileID:      detail.ProfileID.String(),
			IsActive:       detail.IsActive,
		},
	}

	if detail.ManagerID != nil {
		resp.Data.ManagerID = detail.ManagerID.String()
	}

	resp.Data.OrgNode = &struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
	}{
		ID:   detail.OrgNodeID.String(),
		Name: detail.OrgNodeName,
		Path: detail.OrgNodePath,
	}

	if detail.ManagerName != "" {
		parts := strings.SplitN(detail.ManagerName, " ", 2)
		mgr := &struct {
			ID        string `json:"id"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}{
			ID:        detail.ManagerID.String(),
			FirstName: parts[0],
		}
		if len(parts) > 1 {
			mgr.LastName = parts[1]
		}
		resp.Data.Manager = mgr
	}

	return resp
}

// toString converts an interface{} to string safely.
func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
