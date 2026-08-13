package goal

import (
	"context"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// GlobalGoalServicer handles global goal business logic.
type GlobalGoalServicer interface {
	CreateGlobalGoal(ctx context.Context, req CreateGlobalGoalRequest) (*repogoal.GlobalGoalRow, error)
	GetGlobalGoal(ctx context.Context, goalID uuid.UUID) (*repogoal.GlobalGoalRow, error)
	ListGlobalGoals(ctx context.Context, cycleID uuid.UUID) ([]*repogoal.GlobalGoalRow, error)
	UpdateGlobalGoal(ctx context.Context, goalID uuid.UUID, req UpdateGlobalGoalRequest) (*repogoal.GlobalGoalRow, error)
	DeleteGlobalGoal(ctx context.Context, goalID uuid.UUID) error
	ExecuteRules(ctx context.Context, goalID uuid.UUID) (int, error)
}

// CreateGlobalGoalRequest is the request body for creating a global goal.
type CreateGlobalGoalRequest struct {
	Name        string                    `json:"name" validate:"required"`
	Description string                    `json:"description"`
	Unit        string                    `json:"unit" validate:"required,oneof=porcentaje moneda numero"`
	Direction   string                    `json:"direction" validate:"required,oneof=ascendente descendente"`
	GoalKind    string                    `json:"goal_kind" validate:"required,oneof=qualitative quantitative"`
	Weight      float64                   `json:"weight" validate:"required,min=0,max=100"`
	TargetValue float64                   `json:"target_value" validate:"required,gt=0"`
	Assignments []CreateAssignmentRequest `json:"assignments"`
	Rules       []CreateRuleRequest       `json:"rules"`
}

// CreateAssignmentRequest is the request body for creating an assignment.
type CreateAssignmentRequest struct {
	EmployeeID    uuid.UUID `json:"employee_id" validate:"required"`
	Weight        float64   `json:"weight" validate:"required,min=0,max=100"`
	TargetValue   float64   `json:"target_value" validate:"required,gt=0"`
	BaselineValue *float64  `json:"baseline_value,omitempty"`
}

// CreateRuleRequest is the request body for creating a rule.
type CreateRuleRequest struct {
	RuleType         string     `json:"rule_type" validate:"required,oneof=department min_direct_reports role"`
	DepartmentID     *uuid.UUID `json:"department_id,omitempty"`
	MinDirectReports *int       `json:"min_direct_reports,omitempty"`
	ProfileID        *uuid.UUID `json:"profile_id,omitempty"`
	DefaultWeight    float64    `json:"default_weight" validate:"required,min=0,max=100"`
	DefaultTarget    float64    `json:"default_target" validate:"omitempty,gt=0"`
}

// UpdateGlobalGoalRequest is the request body for updating a global goal.
type UpdateGlobalGoalRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Unit        string  `json:"unit" validate:"required,oneof=porcentaje moneda numero"`
	Direction   string  `json:"direction" validate:"required,oneof=ascendente descendente"`
	GoalKind    string  `json:"goal_kind" validate:"required,oneof=qualitative quantitative"`
	Weight      float64 `json:"weight" validate:"required,min=0,max=100"`
	TargetValue float64 `json:"target_value" validate:"required,gt=0"`
}

// globalGoalService implements GlobalGoalServicer.
type globalGoalService struct {
	repo *repogoal.GlobalGoalRepo
}

// NewGlobalGoalService creates a new GlobalGoalService.
func NewGlobalGoalService(repo *repogoal.GlobalGoalRepo) GlobalGoalServicer {
	return &globalGoalService{repo: repo}
}

// CreateGlobalGoal creates a new global goal.
func (s *globalGoalService) CreateGlobalGoal(ctx context.Context, req CreateGlobalGoalRequest) (*repogoal.GlobalGoalRow, error) {
	// Get current user from context
	userID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return nil, pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}

	// Convert assignments
	assignments := make([]*repogoal.GlobalAssignmentRow, 0, len(req.Assignments))
	for _, a := range req.Assignments {
		assignments = append(assignments, &repogoal.GlobalAssignmentRow{
			EmployeeID:    a.EmployeeID,
			Weight:        a.Weight,
			TargetValue:   a.TargetValue,
			BaselineValue: a.BaselineValue,
		})
	}

	// Convert rules
	rules := make([]*repogoal.GlobalRuleRow, 0, len(req.Rules))
	for _, r := range req.Rules {
		defaultTarget := r.DefaultTarget
		if defaultTarget == 0 {
			defaultTarget = 100
		}
		rules = append(rules, &repogoal.GlobalRuleRow{
			RuleType:         r.RuleType,
			DepartmentID:     r.DepartmentID,
			MinDirectReports: r.MinDirectReports,
			ProfileID:        r.ProfileID,
			DefaultWeight:    r.DefaultWeight,
			DefaultTarget:    defaultTarget,
		})
	}

	return s.repo.CreateGlobalGoal(ctx, uuid.Nil, userID, req.Name, req.Description, req.Unit, req.Direction, req.GoalKind, req.Weight, req.TargetValue, assignments, rules)
}

// GetGlobalGoal retrieves a global goal.
func (s *globalGoalService) GetGlobalGoal(ctx context.Context, goalID uuid.UUID) (*repogoal.GlobalGoalRow, error) {
	return s.repo.GetGlobalGoal(ctx, goalID)
}

// ListGlobalGoals lists all global goals.
func (s *globalGoalService) ListGlobalGoals(ctx context.Context, cycleID uuid.UUID) ([]*repogoal.GlobalGoalRow, error) {
	return s.repo.ListGlobalGoalsByCycle(ctx, cycleID)
}

// UpdateGlobalGoal updates a global goal.
func (s *globalGoalService) UpdateGlobalGoal(ctx context.Context, goalID uuid.UUID, req UpdateGlobalGoalRequest) (*repogoal.GlobalGoalRow, error) {
	return s.repo.UpdateGlobalGoal(ctx, goalID, req.Name, req.Description, req.Unit, req.Direction, req.GoalKind, req.Weight, req.TargetValue, 0)
}

// DeleteGlobalGoal deletes a global goal.
func (s *globalGoalService) DeleteGlobalGoal(ctx context.Context, goalID uuid.UUID) error {
	return s.repo.DeleteGlobalGoal(ctx, goalID)
}

// ExecuteRules executes mass assignment rules.
func (s *globalGoalService) ExecuteRules(ctx context.Context, goalID uuid.UUID) (int, error) {
	return s.repo.ExecuteRules(ctx, goalID)
}
