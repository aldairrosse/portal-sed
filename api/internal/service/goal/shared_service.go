package goal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// SharedGoalServicer handles shared goal business logic.
type SharedGoalServicer interface {
	CreateSharedGoal(ctx context.Context, req CreateSharedGoalRequest) (*repogoal.SharedGoalRow, error)
	GetSharedGoal(ctx context.Context, goalID uuid.UUID) (*repogoal.SharedGoalRow, error)
	ListSharedGoalsAsCreator(ctx context.Context) ([]*repogoal.SharedGoalRow, error)
	ListSharedGoalsAsMember(ctx context.Context) ([]*repogoal.SharedGoalRow, error)
	UpdateSharedGoal(ctx context.Context, goalID uuid.UUID, req UpdateSharedGoalRequest) (*repogoal.SharedGoalRow, error)
	DeleteSharedGoal(ctx context.Context, goalID uuid.UUID) error
	AddMember(ctx context.Context, goalID uuid.UUID, req AddMemberRequest) (*repogoal.SharedMemberRow, error)
	RemoveMember(ctx context.Context, goalID, employeeID uuid.UUID) error
	UpdateProgress(ctx context.Context, goalID, employeeID uuid.UUID, req UpdateProgressRequest) error
}

// CreateSharedGoalRequest is the request body for creating a shared goal.
type CreateSharedGoalRequest struct {
	Name             string                `json:"name" validate:"required"`
	Description      string                `json:"description"`
	Unit             string                `json:"unit" validate:"required,oneof=porcentaje moneda numero"`
	Direction        string                `json:"direction" validate:"required,oneof=ascendente descendente"`
	GoalKind         string                `json:"goal_kind" validate:"required,oneof=qualitative quantitative"`
	Weight           float64               `json:"weight" validate:"required,min=0,max=100"`
	TargetValue      float64               `json:"target_value" validate:"required,gt=0"`
	GroupName        string                `json:"group_name" validate:"required"`
	GroupDescription string                `json:"group_description"`
	Members          []CreateMemberRequest `json:"members" validate:"required,min=1"`
}

// CreateMemberRequest is the request body for creating a member.
type CreateMemberRequest struct {
	EmployeeID    uuid.UUID `json:"employee_id" validate:"required"`
	Weight        float64   `json:"weight" validate:"required,min=0,max=100"`
	TargetValue   float64   `json:"target_value" validate:"required,gt=0"`
	BaselineValue *float64  `json:"baseline_value,omitempty"`
}

// UpdateSharedGoalRequest is the request body for updating a shared goal.
type UpdateSharedGoalRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Unit        string  `json:"unit" validate:"required,oneof=porcentaje moneda numero"`
	Direction   string  `json:"direction" validate:"required,oneof=ascendente descendente"`
	GoalKind    string  `json:"goal_kind" validate:"required,oneof=qualitative quantitative"`
	Weight      float64 `json:"weight" validate:"required,min=0,max=100"`
	TargetValue float64 `json:"target_value" validate:"required,gt=0"`
}

// AddMemberRequest is the request body for adding a member.
type AddMemberRequest struct {
	EmployeeID    uuid.UUID `json:"employee_id" validate:"required"`
	Weight        float64   `json:"weight" validate:"required,min=0,max=100"`
	TargetValue   float64   `json:"target_value" validate:"required,gt=0"`
	BaselineValue *float64  `json:"baseline_value,omitempty"`
}

// UpdateProgressRequest is the request body for updating progress.
type UpdateProgressRequest struct {
	CurrentValue float64 `json:"current_value" validate:"required,min=0"`
}

// sharedGoalService implements SharedGoalServicer.
type sharedGoalService struct {
	repo *repogoal.SharedGoalRepo
}

// NewSharedGoalService creates a new SharedGoalService.
func NewSharedGoalService(repo *repogoal.SharedGoalRepo) SharedGoalServicer {
	return &sharedGoalService{repo: repo}
}

// CreateSharedGoal creates a new shared goal.
func (s *sharedGoalService) CreateSharedGoal(ctx context.Context, req CreateSharedGoalRequest) (*repogoal.SharedGoalRow, error) {
	userID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return nil, pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}

	members := make([]*repogoal.SharedMemberRow, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, &repogoal.SharedMemberRow{
			EmployeeID:    m.EmployeeID,
			Weight:        m.Weight,
			TargetValue:   m.TargetValue,
			BaselineValue: m.BaselineValue,
		})
	}

	return s.repo.CreateSharedGoal(ctx, userID, req.Name, req.Description, req.Unit, req.Direction, req.GoalKind, req.Weight, req.TargetValue, req.GroupName, req.GroupDescription, members)
}

// GetSharedGoal retrieves a shared goal.
func (s *sharedGoalService) GetSharedGoal(ctx context.Context, goalID uuid.UUID) (*repogoal.SharedGoalRow, error) {
	return s.repo.GetSharedGoal(ctx, goalID)
}

// ListSharedGoalsAsCreator lists shared goals created by the user.
func (s *sharedGoalService) ListSharedGoalsAsCreator(ctx context.Context) ([]*repogoal.SharedGoalRow, error) {
	userID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return nil, pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}
	return s.repo.ListSharedGoalsAsCreator(ctx, userID)
}

// ListSharedGoalsAsMember lists shared goals where the user is a member.
func (s *sharedGoalService) ListSharedGoalsAsMember(ctx context.Context) ([]*repogoal.SharedGoalRow, error) {
	userID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return nil, pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}
	return s.repo.ListSharedGoalsAsMember(ctx, userID)
}

// UpdateSharedGoal updates a shared goal.
func (s *sharedGoalService) UpdateSharedGoal(ctx context.Context, goalID uuid.UUID, req UpdateSharedGoalRequest) (*repogoal.SharedGoalRow, error) {
	goal, err := s.repo.GetSharedGoal(ctx, goalID)
	if err != nil {
		return nil, err
	}
	userID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return nil, pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}
	if goal.CreatedBy != userID {
		return nil, ErrNotCreator
	}

	return s.repo.UpdateSharedGoal(ctx, goalID, req.Name, req.Description, req.Unit, req.Direction, req.GoalKind, req.Weight, req.TargetValue)
}

// DeleteSharedGoal deletes a shared goal.
func (s *sharedGoalService) DeleteSharedGoal(ctx context.Context, goalID uuid.UUID) error {
	goal, err := s.repo.GetSharedGoal(ctx, goalID)
	if err != nil {
		return err
	}
	userID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}
	if goal.CreatedBy != userID {
		return ErrNotCreator
	}

	return s.repo.DeleteSharedGoal(ctx, goalID)
}

// AddMember adds a member to a shared goal.
func (s *sharedGoalService) AddMember(ctx context.Context, goalID uuid.UUID, req AddMemberRequest) (*repogoal.SharedMemberRow, error) {
	goal, err := s.repo.GetSharedGoal(ctx, goalID)
	if err != nil {
		return nil, err
	}
	userID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return nil, pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}
	if goal.CreatedBy != userID {
		return nil, ErrNotCreator
	}

	return s.repo.AddMember(ctx, goalID, req.EmployeeID, req.Weight, req.TargetValue, req.BaselineValue)
}

// RemoveMember removes a member from a shared goal.
func (s *sharedGoalService) RemoveMember(ctx context.Context, goalID, employeeID uuid.UUID) error {
	goal, err := s.repo.GetSharedGoal(ctx, goalID)
	if err != nil {
		return err
	}
	userID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}
	if goal.CreatedBy != userID {
		return ErrNotCreator
	}

	return s.repo.RemoveMember(ctx, goalID, employeeID)
}

// UpdateProgress updates progress for a member.
func (s *sharedGoalService) UpdateProgress(ctx context.Context, goalID, employeeID uuid.UUID, req UpdateProgressRequest) error {
	goal, err := s.repo.GetSharedGoal(ctx, goalID)
	if err != nil {
		return err
	}
	userID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}
	if goal.CreatedBy != userID {
		return ErrNotCreator
	}

	return s.repo.UpdateProgress(ctx, goalID, employeeID, req.CurrentValue)
}

// ErrNotCreator is returned when the user is not the creator of the shared goal.
var ErrNotCreator = fmt.Errorf("only the creator can modify this shared goal")
