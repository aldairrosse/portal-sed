package org

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/org"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
)

// MetricsService defines the interface for area metrics operations.
type MetricsService interface {
	GetAreaMetrics(ctx context.Context, nodeID, cycleID string) (*dto.AreaMetricsResponse, error)
}

type metricsService struct {
	metricsRepo *repo.MetricsRepo
	nodeRepo    *repo.OrgNodeRepo
	client      *internal.Client
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(metricsRepo *repo.MetricsRepo, nodeRepo *repo.OrgNodeRepo, client *internal.Client) MetricsService {
	return &metricsService{
		metricsRepo: metricsRepo,
		nodeRepo:    nodeRepo,
		client:      client,
	}
}

// GetAreaMetrics computes aggregated metrics for the employees of a given org node.
func (s *metricsService) GetAreaMetrics(ctx context.Context, nodeID, cycleID string) (*dto.AreaMetricsResponse, error) {
	// Parse and validate nodeId
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "nodeId must be a valid UUID", err)
	}

	// Verify node exists
	_, err = s.nodeRepo.GetByID(ctx, nodeUUID)
	if err != nil {
		if err == repo.ErrNodeNotFound {
			return nil, errors.NewDomainError(errors.NodeNotFound, "Org node not found", err)
		}
		return nil, err
	}

	// Get direct employees
	employees, err := s.metricsRepo.GetDirectEmployees(ctx, nodeUUID)
	if err != nil {
		return nil, err
	}

	// Build response
	resp := &dto.AreaMetricsResponse{
		NodeID:        nodeID,
		EmployeeCount: len(employees),
	}

	if len(employees) == 0 {
		resp.Employees = []dto.AreaMetricsEmployee{}
		return resp, nil
	}

	// Extract employee IDs
	empIDs := make([]uuid.UUID, len(employees))
	for i, emp := range employees {
		empIDs[i] = emp.ID
	}

	// Build employees list for response
	resp.Employees = make([]dto.AreaMetricsEmployee, len(employees))
	for i, emp := range employees {
		resp.Employees[i] = dto.AreaMetricsEmployee{
			ID:        emp.ID.String(),
			FirstName: emp.FirstName,
			LastName:  emp.LastName,
			ProfileID: emp.ProfileID.String(),
		}
	}

	// Get goals for those employees
	goals, err := s.metricsRepo.GetGoalsByEmployees(ctx, empIDs)
	if err != nil {
		return nil, err
	}

	if len(goals) > 0 {
		resp.AvgProgress = computeAvgProgress(goals)
		resp.CompletedGoals = countCompleted(goals)
		resp.PendingGoals = countPending(goals)
		resp.EmployeesWithGoals = countEmployeesWithGoals(goals)
	}

	// Parse cycle ID if provided
	var cycleUUID uuid.UUID
	if cycleID != "" {
		cycleUUID, err = uuid.Parse(cycleID)
		if err != nil {
			return nil, errors.NewDomainError(errors.InvalidRequest, "cycleId must be a valid UUID", err)
		}
	}

	// Get RH evaluations if cycleID is provided
	if cycleUUID != uuid.Nil {
		ratings, err := s.metricsRepo.GetRHEvaluationsByEmployees(ctx, empIDs, cycleUUID)
		if err != nil {
			return nil, err
		}

		if len(ratings) > 0 {
			resp.AvgRating = computeAvgRating(ratings)
			resp.RatingsCount = len(ratings)
		}
	}

	return resp, nil
}

// computeAvgProgress calculates the mean progress percentage across all goals.
// Progress for each goal is (current_value / target_value) * 100.
// Returns nil if no goals are provided.
func computeAvgProgress(goals []*repo.GoalRow) *float64 {
	if len(goals) == 0 {
		return nil
	}
	var sum float64
	count := 0
	for _, g := range goals {
		if g.TargetValue > 0 {
			sum += (g.CurrentValue / g.TargetValue) * 100
			count++
		}
	}
	if count == 0 {
		return nil
	}
	avg := sum / float64(count)
	// Round to one decimal place
	avg = math.Round(avg*10) / 10
	return &avg
}

// computeAvgRating calculates the mean of all RH ratings.
// Returns nil if no ratings are provided.
func computeAvgRating(ratings []*repo.RHRatingRow) *float64 {
	if len(ratings) == 0 {
		return nil
	}
	var sum float64
	for _, r := range ratings {
		sum += r.RHRating
	}
	avg := sum / float64(len(ratings))
	// Round to one decimal place
	avg = math.Round(avg*10) / 10
	return &avg
}

// countCompleted counts goals where current_value >= target_value.
func countCompleted(goals []*repo.GoalRow) int {
	count := 0
	for _, g := range goals {
		if g.CurrentValue >= g.TargetValue {
			count++
		}
	}
	return count
}

// countPending counts goals where current_value < target_value.
func countPending(goals []*repo.GoalRow) int {
	count := 0
	for _, g := range goals {
		if g.CurrentValue < g.TargetValue {
			count++
		}
	}
	return count
}

// countEmployeesWithGoals returns the count of distinct employees that have
// at least one goal.
func countEmployeesWithGoals(goals []*repo.GoalRow) int {
	seen := make(map[uuid.UUID]struct{})
	for _, g := range goals {
		seen[g.EmployeeID] = struct{}{}
	}
	return len(seen)
}
