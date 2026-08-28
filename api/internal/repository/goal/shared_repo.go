package goal

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/sharedgoalgroup"
	"github.com/sed-evaluacion-desempeno/api/internal/sharedgoalmember"
)

// deleteSharedDeps removes members and groups for a goal inside the given transaction.
func deleteSharedDeps(ctx context.Context, tx *internal.Tx, goalID uuid.UUID) error {
	groupIDs, err := tx.SharedGoalGroup.Query().Where(sharedgoalgroup.GoalID(goalID)).IDs(ctx)
	if err != nil {
		return fmt.Errorf("query shared groups: %w", err)
	}
	if len(groupIDs) > 0 {
		if _, err := tx.SharedGoalMember.Delete().Where(sharedgoalmember.GroupIDIn(groupIDs...)).Exec(ctx); err != nil {
			return fmt.Errorf("delete shared members: %w", err)
		}
	}
	if _, err := tx.SharedGoalGroup.Delete().Where(sharedgoalgroup.GoalID(goalID)).Exec(ctx); err != nil {
		return fmt.Errorf("delete shared groups: %w", err)
	}
	return nil
}

// SharedGoalRow is the full representation of a shared goal with its group and members.
type SharedGoalRow struct {
	ID           uuid.UUID          `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Unit         string             `json:"unit"`
	Direction    string             `json:"direction"`
	Weight       float64            `json:"weight"`
	TargetValue  float64            `json:"target_value"`
	BaselineValue *float64          `json:"baseline_value,omitempty"`
	CurrentValue float64            `json:"current_value"`
	GoalKind     string             `json:"goal_kind"`
	State        string             `json:"state"`
	CreatedBy    uuid.UUID          `json:"created_by"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Group        *SharedGroupRow    `json:"group"`
	Members      []*SharedMemberRow `json:"members"`
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

// CreateSharedGoal creates a new shared goal with its group and members atomically.
func (r *SharedGoalRepo) CreateSharedGoal(ctx context.Context, createdBy uuid.UUID, name, description, unit, direction, goalKind string, weight, targetValue float64, baselineValue *float64, groupName, groupDescription string, members []*SharedMemberRow) (*SharedGoalRow, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	g, err := tx.Goal.Create().
		SetName(name).
		SetDescription(description).
		SetUnit(goal.Unit(unit)).
		SetDirection(goal.Direction(direction)).
		SetWeight(weight).
		SetTargetValue(targetValue).
		SetNillableBaselineValue(baselineValue).
		SetGoalKind(goal.GoalKind(goalKind)).
		SetState(goal.StateBorrador).
		SetNillableCategoryID(nil).
		SetCreatedBy(createdBy).
		SetUpdatedBy(createdBy).
		SetType(goal.TypeShared).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	grp, err := tx.SharedGoalGroup.Create().
		SetGoalID(g.ID).
		SetCreatedBy(createdBy).
		SetUpdatedBy(createdBy).
		SetName(groupName).
		SetDescription(groupDescription).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	for _, m := range members {
		create := tx.SharedGoalMember.Create().
			SetGroupID(grp.ID).
			SetEmployeeID(m.EmployeeID).
			SetWeight(m.Weight).
			SetTargetValue(m.TargetValue)
		if m.BaselineValue != nil {
			create = create.SetBaselineValue(*m.BaselineValue)
		}
		if _, err := create.Save(ctx); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
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
		ID:            g.ID,
		Name:          g.Name,
		Description:   g.Description,
		Unit:          string(g.Unit),
		Direction:     string(g.Direction),
		Weight:        g.Weight,
		TargetValue:   g.TargetValue,
		BaselineValue: g.BaselineValue,
		CurrentValue:  g.CurrentValue,
		GoalKind:      goalKindValue(g.GoalKind),
		State:         string(g.State),
		CreatedBy:     g.CreatedBy,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
		Members:       make([]*SharedMemberRow, 0),
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
			q.WithMembers(func(mq *internal.SharedGoalMemberQuery) {
				mq.Order(internal.Asc(sharedgoalmember.FieldEmployeeID))
			})
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
func (r *SharedGoalRepo) UpdateSharedGoal(ctx context.Context, goalID uuid.UUID, name, description, unit, direction, goalKind string, weight, targetValue float64, currentValue *float64, baselineValue *float64) (*SharedGoalRow, error) {
	update := r.client.Goal.UpdateOneID(goalID).
		SetName(name).
		SetDescription(description).
		SetUnit(goal.Unit(unit)).
		SetDirection(goal.Direction(direction)).
		SetGoalKind(goal.GoalKind(goalKind)).
		SetWeight(weight).
		SetTargetValue(targetValue)

	if currentValue != nil {
		update = update.SetCurrentValue(*currentValue)
	}
	if baselineValue != nil {
		update = update.SetBaselineValue(*baselineValue)
	}

	_, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.GetSharedGoal(ctx, goalID)
}

// SyncSharedGoalMembers replaces members for a shared goal's group atomically (delete + recreate).
func (r *SharedGoalRepo) SyncSharedGoalMembers(ctx context.Context, goalID uuid.UUID, members []*SharedMemberRow) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	grp, err := tx.SharedGoalGroup.Query().Where(sharedgoalgroup.GoalID(goalID)).Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.SharedGoalMember.Delete().Where(sharedgoalmember.GroupID(grp.ID)).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete old members: %w", err)
	}
	for _, m := range members {
		create := tx.SharedGoalMember.Create().
			SetGroupID(grp.ID).
			SetEmployeeID(m.EmployeeID).
			SetWeight(m.Weight).
			SetTargetValue(m.TargetValue)
		if m.BaselineValue != nil {
			create = create.SetBaselineValue(*m.BaselineValue)
		}
		if _, err := create.Save(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// DeleteSharedGoal deletes a shared goal and its group/members.
func (r *SharedGoalRepo) DeleteSharedGoal(ctx context.Context, goalID uuid.UUID) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	if err := deleteSharedDeps(ctx, tx, goalID); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Goal.DeleteOneID(goalID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		if internal.IsNotFound(err) {
			return pkgerrors.ErrGoalNotFound
		}
		return err
	}
	return tx.Commit()
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
			ID:            g.ID,
			Name:          g.Name,
			Description:   g.Description,
			Unit:          string(g.Unit),
			Direction:     string(g.Direction),
			Weight:        g.Weight,
			TargetValue:   g.TargetValue,
			BaselineValue: g.BaselineValue,
			CurrentValue:  g.CurrentValue,
			GoalKind:      goalKindValue(g.GoalKind),
			State:         string(g.State),
			CreatedBy:     g.CreatedBy,
			CreatedAt:     g.CreatedAt,
			UpdatedAt:     g.UpdatedAt,
			Members:       make([]*SharedMemberRow, 0),
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
		rows = append(rows, row)
	}
	return rows
}
