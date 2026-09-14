package org

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/dto/org"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/state"
	repocycle "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
	repoeval "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
)

// EmployeeService defines the interface for employee operations.
type EmployeeService interface {
	ListEmployees(ctx context.Context, treeID, nodeID, profileID, isActive, query string, offset, limit int) (*org.EmployeeListResponse, error)
	GetEmployee(ctx context.Context, empID string) (*org.EmployeeDetailResponse, error)
	SearchEmployees(ctx context.Context, query string, limit int) (*org.EmployeeListResponse, error)
	UpdateEmployee(ctx context.Context, empID string, req org.UpdateEmployeeRequest, updatedBy uuid.UUID) (*org.EmployeeDetailResponse, error)
}

type employeeService struct {
	empRepo   *repo.EmployeeRepo
	client    *internal.Client
	cycleRepo *repocycle.CycleRepo
	evalRepo  *repoeval.EvaluationRepo
}

// AttachEmployeeEvaluationDeps enables RH list enrichment (active cycle +
// avg/goal batches). No-op on type mismatch; nil repos = fallback (fields
// stay null/omitted). Keeps NewEmployeeService signature stable for tests.
func AttachEmployeeEvaluationDeps(svc EmployeeService, cycleRepo *repocycle.CycleRepo, evalRepo *repoeval.EvaluationRepo) {
	if s, ok := svc.(*employeeService); ok {
		s.cycleRepo = cycleRepo
		s.evalRepo = evalRepo
	}
}
// NewEmployeeService creates a new EmployeeService.
func NewEmployeeService(empRepo *repo.EmployeeRepo, client *internal.Client) EmployeeService {
	return &employeeService{
		empRepo: empRepo,
		client:  client,
	}
}

func (s *employeeService) ListEmployees(ctx context.Context, treeID, nodeID, profileID, isActive, query string, offset, limit int) (*org.EmployeeListResponse, error) {
	filter := repo.EmployeeFilter{
		Query:  query,
		Offset: offset,
		Limit:  limit,
	}

	if treeID != "" {
		id, err := uuid.Parse(treeID)
		if err == nil {
			filter.TreeID = &id
		}
	}
	if nodeID != "" {
		id, err := uuid.Parse(nodeID)
		if err == nil {
			filter.NodeID = &id
		}
	}
	if profileID != "" {
		id, err := uuid.Parse(profileID)
		if err == nil {
			filter.ProfileID = &id
		}
	}
	if isActive != "" {
		active := strings.EqualFold(isActive, "true") || isActive == "1"
		filter.IsActive = &active
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	} else if filter.Limit > 200 {
		filter.Limit = 200
	}

	rows, err := s.empRepo.ListWithProfiles(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := &org.EmployeeListResponse{
		Data: make([]org.EmployeeListItem, len(rows)),
	}
	resp.Meta.Limit = filter.Limit
	resp.Meta.Offset = filter.Offset

	// Total count (same filters, no pagination)
	countFilter := filter
	countFilter.Offset = 0
	countFilter.Limit = 0
	total, err := s.empRepo.CountWithProfiles(ctx, countFilter)
	if err == nil {
		resp.Meta.Total = total
		resp.Meta.HasMore = filter.Offset+len(rows) < total
	}

	// RH enrichment (mirror GetMyEvaluateesPaginated 167-256): active cycle
	// -> phase/phaseKind -> BatchAvg + BatchGoalStatus + CompetencyTotalCount
	// in 3 queries (no N+1). Nil-safe: any missing repo/error => plain items.
	var targetCycleID *uuid.UUID
	activePhase := ""
	if s.cycleRepo != nil {
		if active, err := s.cycleRepo.GetActive(ctx); err == nil && active != nil {
			targetCycleID = &active.ID
			activePhase = string(active.CurrentPhase)
		}
	}
	var empIDs []uuid.UUID
	if targetCycleID != nil && activePhase != "" && len(rows) > 0 && s.evalRepo != nil {
		empIDs = make([]uuid.UUID, len(rows))
		for i, r := range rows {
			empIDs[i] = r.ID
		}
	}
	avgMap := map[uuid.UUID]repoeval.EmployeeAvg{}
	goalMap := map[uuid.UUID]repoeval.EmployeeGoalStatus{}
	compTotal := 0
	if targetCycleID != nil && activePhase != "" && len(empIDs) > 0 && s.evalRepo != nil {
		if m, err := s.evalRepo.BatchAvgByEmployeeIDs(ctx, *targetCycleID, activePhase, empIDs); err == nil {
			avgMap = m
		}
		if m, err := s.evalRepo.BatchGoalStatusByEmployeeIDs(ctx, *targetCycleID, activePhase, empIDs); err == nil {
			goalMap = m
		}
		if n, err := s.evalRepo.CompetencyTotalCount(ctx); err == nil {
			compTotal = n
		}
	}
	phaseKind := ""
	normPhase := ""
	if strings.TrimSpace(activePhase) != "" {
		normPhase = state.NormalizePhase(activePhase)
		if state.IsMidYearPhase(activePhase) {
			phaseKind = "avance"
		} else {
			phaseKind = "cierre"
		}
	}

	for i, r := range rows {
		item := employeeRowToItem(r)
		if a, ok := avgMap[r.ID]; ok {
			item.SelfAvg = a.SelfAvg
			item.RhAvg = a.RhAvg
		}
		if targetCycleID != nil && phaseKind != "" {
			rated := 0
			if a, ok := avgMap[r.ID]; ok {
				rated = a.SelfCount
			}
			gt, gd := 0, 0
			if g, ok := goalMap[r.ID]; ok {
				gt, gd = g.Total, g.Done
			}
			item.Phase = normPhase
			item.PhaseKind = phaseKind
			if compTotal > 0 {
				tc := compTotal
				item.TotalCompetencies = &tc
			}
			ec := rated
			item.EvaluatedCount = &ec
			item.GoalTotal = &gt
			item.GoalDone = &gd
			switch {
			case gt == 0:
				item.MetasStatusFase = "no_iniciado"
			case gd == 0:
				item.MetasStatusFase = "pending"
			case gd < gt:
				item.MetasStatusFase = "in-progress"
			default:
				item.MetasStatusFase = "completed"
			}
			switch {
			case rated == 0 && gd == 0:
				item.EvaluationStatus = "pending"
			case compTotal <= 0:
				item.EvaluationStatus = "in-progress"
			case rated < compTotal || gd < gt:
				item.EvaluationStatus = "in-progress"
			default:
				item.EvaluationStatus = "completed"
			}
		}
		resp.Data[i] = item
	}

	// Global completed count (RH) — mismo filtro, sin paginación.
	// No afecta a mis-evaluados: este servicio solo sirve GET /employees.
	// Lógica de completada idéntica al per-row: self competencies + metas fase.
	if targetCycleID != nil && activePhase != "" && phaseKind != "" && s.evalRepo != nil && total > 0 {
		if compTotal == 0 {
			if n, err := s.evalRepo.CompetencyTotalCount(ctx); err == nil {
				compTotal = n
			}
		}
		allIDs, err := s.empRepo.ListIDsWithProfiles(ctx, countFilter)
		if err == nil {
			if len(allIDs) == 0 {
				zero := 0
				resp.Meta.CompletedCount = &zero
			} else {
				const chunk = 500
				globalAvg := map[uuid.UUID]repoeval.EmployeeAvg{}
				globalGoals := map[uuid.UUID]repoeval.EmployeeGoalStatus{}
				for sIdx := 0; sIdx < len(allIDs); sIdx += chunk {
					e := sIdx + chunk
					if e > len(allIDs) {
						e = len(allIDs)
					}
					slice := allIDs[sIdx:e]
					if m, err := s.evalRepo.BatchAvgByEmployeeIDs(ctx, *targetCycleID, activePhase, slice); err == nil {
						for k, v := range m {
							globalAvg[k] = v
						}
					}
					if m, err := s.evalRepo.BatchGoalStatusByEmployeeIDs(ctx, *targetCycleID, activePhase, slice); err == nil {
						for k, v := range m {
							globalGoals[k] = v
						}
					}
				}
				completed := 0
				for _, id := range allIDs {
					rated := 0
					if a, ok := globalAvg[id]; ok {
						rated = a.SelfCount
					}
					gt, gd := 0, 0
					if g, ok := globalGoals[id]; ok {
						gt, gd = g.Total, g.Done
					}
					var status string
					switch {
					case rated == 0 && gd == 0:
						status = "pending"
					case compTotal <= 0:
						status = "in-progress"
					case rated < compTotal || gd < gt:
						status = "in-progress"
					default:
						status = "completed"
					}
					if status == "completed" {
						completed++
					}
				}
				resp.Meta.CompletedCount = &completed
			}
		}
	}

	return resp, nil
}

func (s *employeeService) GetEmployee(ctx context.Context, empID string) (*org.EmployeeDetailResponse, error) {
	id, err := uuid.Parse(empID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid employee ID: must be a valid UUID", err)
	}

	detail, err := s.empRepo.GetDetailByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &org.EmployeeDetailResponse{
		Data: org.EmployeeDetail{
			ID:             detail.ID.String(),
			FirstName:      detail.FirstName,
			LastName:       detail.LastName,
			Email:          detail.Email,
			EmployeeNumber: detail.EmployeeNumber,
			OrgNodeID:      detail.OrgNodeID.String(),
			ProfileID:      detail.ProfileID.String(),
			ProfileName:    detail.ProfileName,
			JobTitle:       detail.JobTitle,
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
		resp.Data.Manager = &struct {
			ID        string `json:"id"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}{
			ID:        detail.ManagerID.String(),
			FirstName: strings.Split(detail.ManagerName, " ")[0],
			LastName:  "",
		}
		// Parse manager name
		if parts := strings.SplitN(detail.ManagerName, " ", 2); len(parts) > 1 {
			resp.Data.Manager.FirstName = parts[0]
			resp.Data.Manager.LastName = parts[1]
		} else {
			resp.Data.Manager.FirstName = detail.ManagerName
		}
	}

	return resp, nil
}

func (s *employeeService) SearchEmployees(ctx context.Context, query string, limit int) (*org.EmployeeListResponse, error) {
	if len(strings.TrimSpace(query)) < 2 {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Search query must be at least 2 characters", nil)
	}

	if limit <= 0 {
		limit = 20
	} else if limit > 50 {
		limit = 50
	}

	rows, err := s.empRepo.Search(ctx, strings.TrimSpace(query), limit)
	if err != nil {
		return nil, err
	}

	resp := &org.EmployeeListResponse{
		Data: make([]org.EmployeeListItem, len(rows)),
	}
	resp.Meta.Limit = limit
	resp.Meta.HasMore = false

	for i, r := range rows {
		resp.Data[i] = employeeRowToItem(r)
	}

	return resp, nil
}

func (s *employeeService) UpdateEmployee(ctx context.Context, empID string, req org.UpdateEmployeeRequest, updatedBy uuid.UUID) (*org.EmployeeDetailResponse, error) {
	id, err := uuid.Parse(empID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid employee ID: must be a valid UUID", err)
	}

	profileID, err := uuid.Parse(req.ProfileID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid profileId: must be a valid UUID", err)
	}

	orgNodeID, err := uuid.Parse(req.OrgNodeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid orgNodeId: must be a valid UUID", err)
	}

	exists, err := s.empRepo.ProfileExists(ctx, profileID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Failed to validate profile", err)
	}
	if !exists {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Profile not found", nil)
	}

	exists, err = s.empRepo.OrgNodeExists(ctx, orgNodeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Failed to validate org node", err)
	}
	if !exists {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Org node not found", nil)
	}

	if err := s.empRepo.UpdateProfileAndDepartment(ctx, id, profileID, orgNodeID, updatedBy); err != nil {
		return nil, err
	}

	return s.GetEmployee(ctx, empID)
}

// employeeRowToItem converts an EmployeeRow to an EmployeeListItem.
func employeeRowToItem(r *repo.EmployeeRow) org.EmployeeListItem {
	item := org.EmployeeListItem{
		ID:                 r.ID.String(),
		FirstName:          r.FirstName,
		LastName:           r.LastName,
		Email:              r.Email,
		EmployeeNumber:     r.EmployeeNumber,
		OrgNodeID:          r.OrgNodeID.String(),
		ProfileID:          r.ProfileID.String(),
		ProfileName:        r.ProfileName,
		ProfileDescription: r.ProfileDescription,
		JobTitle:           r.JobTitle,
		IsActive:           r.IsActive,
	}
	if r.ManagerID != nil {
		item.ManagerID = r.ManagerID.String()
	}
	return item
}
