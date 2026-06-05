package goal

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/goal"
	"github.com/sed-evaluacion-desempeno/api/internal/goalcategory"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// GoalRow is the full representation of a Goal including the version field.
type GoalRow struct {
	ID           uuid.UUID  `json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CreatedBy    uuid.UUID  `json:"created_by"`
	UpdatedBy    uuid.UUID  `json:"updated_by"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Unit         string     `json:"unit"`
	Weight       float64    `json:"weight"`
	TargetValue  float64    `json:"target_value"`
	CurrentValue float64    `json:"current_value"`
	State        string     `json:"state"`
	CategoryID   uuid.UUID  `json:"category_id"`
	Version      int        `json:"version"`
}

// goalToRow converts an ent Goal to a GoalRow, fetching the version from the DB.
func (r *GoalRepo) goalToRow(ctx context.Context, g *internal.Goal) (*GoalRow, error) {
	if g == nil {
		return nil, nil
	}
	version, err := r.fetchVersion(ctx, g.ID)
	if err != nil {
		version = 1
	}
	return &GoalRow{
		ID:           g.ID,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		CreatedBy:    g.CreatedBy,
		UpdatedBy:    g.UpdatedBy,
		Name:         g.Name,
		Description:  g.Description,
		Unit:         string(g.Unit),
		Weight:       g.Weight,
		TargetValue:  g.TargetValue,
		CurrentValue: g.CurrentValue,
		State:        string(g.State),
		CategoryID:   g.CategoryID,
		Version:      version,
	}, nil
}

// GoalRepo provides Ent-backed CRUD operations for Goals.
type GoalRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewGoalRepo creates a new GoalRepo.
func NewGoalRepo(client *internal.Client, db *sql.DB) *GoalRepo {
	return &GoalRepo{client: client, db: db}
}

// CreateGoal inserts a new goal with version=1 and state='borrador'.
// Uses raw SQL to set the initial version.
func (r *GoalRepo) CreateGoal(ctx context.Context, catID uuid.UUID, name, description, unit string, weight, targetValue float64) (*GoalRow, error) {
	now := time.Now()
	id := uuid.New()
	createdBy := uuid.Nil // TODO(auth:C7): inject from context
	state := "borrador"

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO goals (id, created_at, updated_at, created_by, updated_by, name, description, unit, weight, target_value, current_value, state, category_id, version)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		id, now, now, createdBy, createdBy, name, description, unit, weight, targetValue, 0.0, state, catID, 1,
	)
	if err != nil {
		return nil, err
	}

	return &GoalRow{
		ID:           id,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		Name:         name,
		Description:  description,
		Unit:         unit,
		Weight:       weight,
		TargetValue:  targetValue,
		CurrentValue: 0,
		State:        state,
		CategoryID:   catID,
		Version:      1,
	}, nil
}

// GetGoal retrieves a single goal by ID.
func (r *GoalRepo) GetGoal(ctx context.Context, goalID uuid.UUID) (*GoalRow, error) {
	g, err := r.client.Goal.Query().
		Where(goal.ID(goalID)).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.ErrGoalNotFound
		}
		return nil, err
	}
	return r.goalToRow(ctx, g)
}

// UpdateGoal updates goal fields with optimistic locking via version.
// Uses raw SQL to atomically check and increment version.
func (r *GoalRepo) UpdateGoal(ctx context.Context, goalID uuid.UUID, name, description, unit string, weight, targetValue float64, expectedVersion int) (*GoalRow, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE goals
		 SET name = $1, description = $2, unit = $3, weight = $4, target_value = $5,
		     updated_at = $6, version = version + 1
		 WHERE id = $7 AND version = $8`,
		name, description, unit, weight, targetValue, now, goalID, expectedVersion,
	)
	if err != nil {
		return nil, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		// Check if the goal exists (version mismatch vs not found)
		exists, _ := r.client.Goal.Query().Where(goal.ID(goalID)).Exist(ctx)
		if !exists {
			return nil, pkgerrors.ErrGoalNotFound
		}
		return nil, pkgerrors.ErrConcurrentModification
	}

	// Fetch the updated goal via Ent to get computed fields
	g, err := r.client.Goal.Query().Where(goal.ID(goalID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return r.goalToRow(ctx, g)
}

// UpdateGoalCurrentValue updates only the currentValue field and increments version.
func (r *GoalRepo) UpdateGoalCurrentValue(ctx context.Context, goalID uuid.UUID, currentValue float64) (*GoalRow, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE goals
		 SET current_value = $1, updated_at = $2, version = version + 1
		 WHERE id = $3`,
		currentValue, now, goalID,
	)
	if err != nil {
		return nil, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, pkgerrors.ErrGoalNotFound
	}

	g, err := r.client.Goal.Query().Where(goal.ID(goalID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return r.goalToRow(ctx, g)
}

// DeleteGoal removes a goal by ID. KPI links are cascade-deleted by the database.
func (r *GoalRepo) DeleteGoal(ctx context.Context, goalID uuid.UUID) error {
	err := r.client.Goal.DeleteOneID(goalID).Exec(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return pkgerrors.ErrGoalNotFound
		}
		return err
	}
	return nil
}

// ListGoalsByCategory retrieves all goals for a category.
func (r *GoalRepo) ListGoalsByCategory(ctx context.Context, catID uuid.UUID) ([]*GoalRow, error) {
	goals, err := r.client.Goal.Query().
		Where(goal.CategoryID(catID)).
		Order(internal.Asc(goal.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]*GoalRow, len(goals))
	for i, g := range goals {
		row, err := r.goalToRow(ctx, g)
		if err != nil {
			return nil, err
		}
		rows[i] = row
	}
	return rows, nil
}

// GetCategory retrieves a category by ID (needed for ownership checks).
func (r *GoalRepo) GetCategory(ctx context.Context, catID uuid.UUID) (*internal.GoalCategory, error) {
	cat, err := r.client.GoalCategory.Query().
		Where(goalcategory.ID(catID)).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.ErrCategoryNotFound
		}
		return nil, err
	}
	return cat, nil
}

// fetchVersion retrieves the version field for a goal.
func (r *GoalRepo) fetchVersion(ctx context.Context, id uuid.UUID) (int, error) {
	var version int
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(version, 1) FROM goals WHERE id = $1`, id,
	).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
