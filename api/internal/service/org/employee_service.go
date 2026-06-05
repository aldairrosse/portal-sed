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

// EmployeeService defines the interface for employee operations.
type EmployeeService interface {
	ListEmployees(ctx context.Context, treeID, nodeID, profileID, isActive, query, cursor string, limit int) (*org.EmployeeListResponse, error)
	GetEmployee(ctx context.Context, empID string) (*org.EmployeeDetailResponse, error)
	SearchEmployees(ctx context.Context, query string, limit int) (*org.EmployeeListResponse, error)
}

type employeeService struct {
	empRepo *repo.EmployeeRepo
	client  *internal.Client
}

// NewEmployeeService creates a new EmployeeService.
func NewEmployeeService(empRepo *repo.EmployeeRepo, client *internal.Client) EmployeeService {
	return &employeeService{
		empRepo: empRepo,
		client:  client,
	}
}

func (s *employeeService) ListEmployees(ctx context.Context, treeID, nodeID, profileID, isActive, query, cursor string, limit int) (*org.EmployeeListResponse, error) {
	filter := repo.EmployeeFilter{
		Query:  query,
		Cursor: cursor,
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

	rows, err := s.empRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	hasMore := len(rows) > filter.Limit
	if hasMore {
		rows = rows[:filter.Limit]
	}

	resp := &org.EmployeeListResponse{
		Data: make([]org.EmployeeListItem, len(rows)),
	}
	resp.Meta.Limit = filter.Limit
	resp.Meta.HasMore = hasMore

	for i, r := range rows {
		resp.Data[i] = employeeRowToItem(r)
	}

	if hasMore && len(rows) > 0 {
		resp.Meta.NextCursor = rows[len(rows)-1].ID.String()
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

// employeeRowToItem converts an EmployeeRow to an EmployeeListItem.
func employeeRowToItem(r *repo.EmployeeRow) org.EmployeeListItem {
	item := org.EmployeeListItem{
		ID:             r.ID.String(),
		FirstName:      r.FirstName,
		LastName:       r.LastName,
		Email:          r.Email,
		EmployeeNumber: r.EmployeeNumber,
		OrgNodeID:      r.OrgNodeID.String(),
		ProfileID:      r.ProfileID.String(),
		IsActive:       r.IsActive,
	}
	if r.ManagerID != nil {
		item.ManagerID = r.ManagerID.String()
	}
	return item
}
