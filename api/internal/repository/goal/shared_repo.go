package goal

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/goal"
	"github.com/sed-evaluacion-desempeno/api/internal/sharedgoalgroup"
	"github.com/sed-evaluacion-desempeno/api/internal/sharedgoalmember"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// SharedGoalRow is the full representation of a shared goal with its group and members.
type SharedGoalRow struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Unit        string             `json:"unit"`
	Direction   string             `json:"direction"`
	Weight      float64            `json:"weight"`
	TargetValue float64            `json:"target_value"`
	GoalKind    string             `json:"goal_kind"`
	State       string             `json:"state"`
	CreatedBy   uuid.UUID          `json:"created_by"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Group       *SharedGroupRow    `json:"group"`
	Members     []*SharedMemberRow `json:"members"`
}

// SharedGroupRow represents the group for a shared goal.
type SharedGroupRow struct {
	ID          uuid.UUID `json:"id"`
	GoalID      uuid.UUID `json:"goal_id"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// SharedMemberRow represents a member of a shared goal group.
type SharedMemberRow struct {
	ID            uuid.UUID `json:"id"`
	GroupID       uuid.UUID `json:"group_id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	Weight        float64   `json:"weight"`
	TargetValue   float64   `json:"target_value"`
	BaselineValue *float64  `json:"baseline_value,omitempty"`
}

// SharedGoalRepo provides Ent-backed operations for shared goals.
type SharedGoalRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewSharedGoalRepo creates a new SharedGoalRepo.
func NewSharedGoalRepo(client *internal.Client, db *sql.DB) *SharedGoalRepo {
	return &SharedGoalRepo{client: client, db: db}
}

// CreateSharedGoal creates a new shared goal with its group and members.
func (r *SharedGoalRepo) CreateSharedGoal(ctx context.Context, createdBy uuid.UUID, name, description, unit, direction, goalKind string, weight, targetValue float64, groupName, groupDescription string, members []*SharedMemberRow) (*SharedGoalRow, error) {
	// Create the goal
	g, err := r.client.Goal.Create().
		SetName(name).
		SetDescription(description).
		SetUnit(goal.Unit(unit)).
		SetDirection(goal.Direction(direction)).
		SetWeight(weight).
		SetTargetValue(targetValue).
		SetGoalKind(goal.GoalKind(goalKind)).
		SetState(goal.StateBorrador).
		SetCategoryID(uuid.Nil). // Shared goals don't belong to a category
		SetCreatedBy(createdBy).
		SetUpdatedBy(createdBy).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Create the group
	grp, err := r.client.SharedGoalGroup.Create().
		SetGoalID(g.ID).
		SetCreatedBy(createdBy).
		SetUpdatedBy(createdBy).
		SetName(groupName).
		SetDescription(groupDescription).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Create members
	for _, m := range members {
		create := r.client.SharedGoalMember.Create().
			SetGroupID(grp.ID).
			SetEmployeeID(m.EmployeeID).
			SetWeight(m.Weight).
			SetTargetValue(m.TargetValue)
		
		if m.BaselineValue != nil {
			create = create.SetBaselineValue(*m.BaselineValue)
		}
		
		_, err := create.Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	return r.GetSharedGoal(ctx, g.ID)
}

// GetSharedGoal retrieves a shared goal with its group and members.
func (r *SharedGoalRepo) GetSharedGoal(ctx context.Context, goalID uuid.UUID) (*SharedGoalRow, error) {
	g, err := r.client.Goal.Query().
		Where(goal.ID(goalID), goal.TypeEQ(goal.TypeShared)).
		WithSharedGroup(func(q *internal.SharedGoalGroupQuery) {
			q.WithMembers(func(mq *internal.SharedGoalMemberQuery) {
				mq.Order(internal.Asc(sharedgoalmember.FieldEmployeeID))
			})
		}).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.ErrGoalNotFound
		}
		return nil, err
	}

	row := &SharedGoalRow{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Unit:        string(g.Unit),
		Direction:   string(g.Direction),
		Weight:      g.Weight,
		TargetValue: g.TargetValue,
		GoalKind:    string(*g.GoalKind),
		State:       string(g.State),
		CreatedBy:   g.CreatedBy,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		Members:     make([]*SharedMemberRow, 0),
	}

	if len(g.Edges.SharedGroup) > 0 {
		grp := g.Edges.SharedGroup[0]
		row.Group = &SharedGroupRow{
			ID:          grp.ID,
			GoalID:      grp.GoalID,
			CreatedBy:   grp.CreatedBy,
			Name:        grp.Name,
			Description: grp.Description,
		}

		for _, m := range grp.Edges.Members {
			row.Members = append(row.Members, &SharedMemberRow{
				ID:            m.ID,
				GroupID:       m.GroupID,
				EmployeeID:    m.EmployeeID,
				Weight:        m.Weight,
				TargetValue:   m.TargetValue,
				BaselineValue: m.BaselineValue,
			})
		}
	}

	return row, nil
}

// ListSharedGoalsAsCreator lists shared goals created by the user.
func (r *SharedGoalRepo) ListSharedGoalsAsCreator(ctx context.Context, creatorID uuid.UUID) ([]*SharedGoalRow, error) {
	goals, err := r.client.Goal.Query().
		Where(goal.TypeEQ(goal.TypeShared), goal.CreatedBy(creatorID)).
		WithSharedGroup(func(q *internal.SharedGoalGroupQuery) {
			q.WithMembers()
		}).
		Order(internal.Desc(goal.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return r.toRows(goals), nil
}

// ListSharedGoalsAsMember lists shared goals where the user is a member.
func (r *SharedGoalRepo) ListSharedGoalsAsMember(ctx context.Context, employeeID uuid.UUID) ([]*SharedGoalRow, error) {
	// Find groups where employee is a member
	memberGroups, err := r.client.SharedGoalMember.Query().
		Where(sharedgoalmember.EmployeeID(employeeID)).
		WithGroup(func(q *internal.SharedGoalGroupQuery) {
			q.WithGoal()
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	rows := make([]*SharedGoalRow, 0)
	seen := make(map[uuid.UUID]bool)

	for _, mg := range memberGroups {
		if mg.Edges.Group != nil && mg.Edges.Group.Edges.Goal != nil {
			goalID := mg.Edges.Group.Edges.Goal.ID
			if !seen[goalID] {
				seen[goalID] = true
				row, err := r.GetSharedGoal(ctx, goalID)
				if err == nil {
					rows = append(rows, row)
				}
			}
		}
	}

	return rows, nil
}

// UpdateSharedGoal updates a shared goal.
func (r *SharedGoalRepo) UpdateSharedGoal(ctx context.Context, goalID uuid.UUID, name, description, unit, direction, goalKind string, weight, targetValue float64) (*SharedGoalRow, error) {
	_, err := r.client.Goal.UpdateOneID(goalID).
		SetName(name).
		SetDescription(description).
		SetUnit(goal.Unit(unit)).
		SetDirection(goal.Direction(direction)).
		SetGoalKind(goal.GoalKind(goalKind)).
		SetWeight(weight).
		SetTargetValue(targetValue).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.GetSharedGoal(ctx, goalID)
}

// DeleteSharedGoal deletes a shared goal and its group/members.
func (r *SharedGoalRepo) DeleteSharedGoal(ctx context.Context, goalID uuid.UUID) error {
	return r.client.Goal.DeleteOneID(goalID).Exec(ctx)
}

// AddMember adds a member to a shared goal group.
func (r *SharedGoalRepo) AddMember(ctx context.Context, goalID, employeeID uuid.UUID, weight, targetValue float64, baselineValue *float64) (*SharedMemberRow, error) {
	// Get or create group
	grp, err := r.client.SharedGoalGroup.Query().
		Where(sharedgoalgroup.GoalID(goalID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	create := r.client.SharedGoalMember.Create().
		SetGroupID(grp.ID).
		SetEmployeeID(employeeID).
		SetWeight(weight).
		SetTargetValue(targetValue)
	
	if baselineValue != nil {
		create = create.SetBaselineValue(*baselineValue)
	}
	
	m, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return &SharedMemberRow{
		ID:            m.ID,
		GroupID:       m.GroupID,
		EmployeeID:    m.EmployeeID,
		Weight:        m.Weight,
		TargetValue:   m.TargetValue,
		BaselineValue: m.BaselineValue,
	}, nil
}

// RemoveMember removes a member from a shared goal group.
func (r *SharedGoalRepo) RemoveMember(ctx context.Context, goalID, employeeID uuid.UUID) error {
	grp, err := r.client.SharedGoalGroup.Query().
		Where(sharedgoalgroup.GoalID(goalID)).
		Only(ctx)
	if err != nil {
		return err
	}

	_, err = r.client.SharedGoalMember.Delete().
		Where(
			sharedgoalmember.GroupID(grp.ID),
			sharedgoalmember.EmployeeID(employeeID),
		).Exec(ctx)
	return err
}

// UpdateProgress updates the progress for a specific member.
func (r *SharedGoalRepo) UpdateProgress(ctx context.Context, goalID, employeeID uuid.UUID, currentValue float64) error {
	// This would update the goal's current_value for the specific member
	// For now, we'll just update the goal's current_value
	_, err := r.client.Goal.UpdateOneID(goalID).
		SetCurrentValue(currentValue).
		Save(ctx)
	return err
}

// toRows converts Goal entities to SharedGoalRow slices.
func (r *SharedGoalRepo) toRows(goals []*internal.Goal) []*SharedGoalRow {
	rows := make([]*SharedGoalRow, 0, len(goals))
	for _, g := range goals {
		row := &SharedGoalRow{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Unit:        string(g.Unit),
			Direction:   string(g.Direction),
			Weight:      g.Weight,
			TargetValue: g.TargetValue,
			GoalKind:    string(*g.GoalKind),
			State:       string(g.State),
			CreatedBy:   g.CreatedBy,
			CreatedAt:   g.CreatedAt,
			UpdatedAt:   g.UpdatedAt,
			Members:     make([]*SharedMemberRow, 0),
		}
		rows = append(rows, row)
	}
	return rows
}
