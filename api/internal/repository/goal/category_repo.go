// Package goal provides the repository layer for goal-related entities.
package goal

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// CategoryRow is the full representation of a GoalCategory.
type CategoryRow struct {
	ID          uuid.UUID  `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	UpdatedBy   uuid.UUID  `json:"updated_by"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Weight      float64    `json:"weight"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	PillarID    *uuid.UUID `json:"pillar_id,omitempty"`
}

// CategoryRepo provides Ent-backed CRUD operations for GoalCategory.
type CategoryRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewCategoryRepo creates a new CategoryRepo.
func NewCategoryRepo(client *internal.Client, db *sql.DB) *CategoryRepo {
	return &CategoryRepo{client: client, db: db}
}

// ListCategoriesByEmployee retrieves all categories for an employee, ordered by name.
func (r *CategoryRepo) ListCategoriesByEmployee(ctx context.Context, empID uuid.UUID) ([]*CategoryRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, created_at, updated_at, created_by, updated_by, name, COALESCE(description, ''), weight, employee_id, pillar_id
		 FROM goal_categories WHERE employee_id = $1 ORDER BY name`,
		empID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*CategoryRow
	for rows.Next() {
		var cat CategoryRow
		var createdAt, updatedAt sql.NullTime
		var pillarID sql.NullString
		if err := rows.Scan(&cat.ID, &createdAt, &updatedAt, &cat.CreatedBy, &cat.UpdatedBy,
			&cat.Name, &cat.Description, &cat.Weight, &cat.EmployeeID, &pillarID); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			cat.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			cat.UpdatedAt = updatedAt.Time
		}
		if pillarID.Valid {
			u := uuid.MustParse(pillarID.String)
			cat.PillarID = &u
		}
		result = append(result, &cat)
	}
	return result, rows.Err()
}

// CreateCategory inserts a new category.
func (r *CategoryRepo) CreateCategory(ctx context.Context, empID uuid.UUID, name, description string, weight float64, pillarID *uuid.UUID) (*CategoryRow, error) {
	cat, err := r.client.GoalCategory.Create().
		SetEmployeeID(empID).
		SetCreatedBy(empID).
		SetUpdatedBy(empID).
		SetName(name).
		SetDescription(description).
		SetWeight(weight).
		Save(ctx)
	if err != nil {
		if internal.IsConstraintError(err) {
			return nil, pkgerrors.ErrDuplicateCategoryName
		}
		return nil, err
	}
	// Workaround: set pillar_id via raw SQL until Ent is regenerated
	if pillarID != nil {
		_, _ = r.db.ExecContext(ctx,
			`UPDATE goal_categories SET pillar_id = $1, updated_at = now() WHERE id = $2`,
			*pillarID, cat.ID,
		)
	}
	return r.GetCategory(ctx, cat.ID)
}

// UpdateCategory updates an existing category's name, description, and weight.
func (r *CategoryRepo) UpdateCategory(ctx context.Context, catID uuid.UUID, name, description string, weight float64, updatedBy uuid.UUID, pillarID *uuid.UUID) (*CategoryRow, error) {
	cat, err := r.client.GoalCategory.UpdateOneID(catID).
		SetName(name).
		SetDescription(description).
		SetWeight(weight).
		SetUpdatedBy(updatedBy).
		Save(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.ErrCategoryNotFound
		}
		if internal.IsConstraintError(err) {
			return nil, pkgerrors.ErrDuplicateCategoryName
		}
		return nil, err
	}
	// Workaround: set pillar_id via raw SQL until Ent is regenerated
	_, _ = r.db.ExecContext(ctx,
		`UPDATE goal_categories SET pillar_id = $1, updated_at = now() WHERE id = $2`,
		pillarID, cat.ID,
	)
	return r.GetCategory(ctx, cat.ID)
}

// DeleteCategory removes a category by ID.
func (r *CategoryRepo) DeleteCategory(ctx context.Context, catID uuid.UUID) error {
	err := r.client.GoalCategory.DeleteOneID(catID).Exec(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return pkgerrors.ErrCategoryNotFound
		}
		return err
	}
	return nil
}

// GetCategory retrieves a single category by ID.
func (r *CategoryRepo) GetCategory(ctx context.Context, catID uuid.UUID) (*CategoryRow, error) {
	var cat CategoryRow
	var createdAt, updatedAt sql.NullTime
	var pillarID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, created_by, updated_by, name, COALESCE(description, ''), weight, employee_id, pillar_id
		 FROM goal_categories WHERE id = $1`,
		catID,
	).Scan(&cat.ID, &createdAt, &updatedAt, &cat.CreatedBy, &cat.UpdatedBy,
		&cat.Name, &cat.Description, &cat.Weight, &cat.EmployeeID, &pillarID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrCategoryNotFound
		}
		return nil, err
	}
	if createdAt.Valid {
		cat.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		cat.UpdatedAt = updatedAt.Time
	}
	if pillarID.Valid {
		u := uuid.MustParse(pillarID.String)
		cat.PillarID = &u
	}
	return &cat, nil
}

// LockCategory acquires a SELECT FOR UPDATE lock on a category row.
// This serialises weight-sum calculations for that category.
// Uses raw SQL because Ent's query builder does not expose ForUpdate().
func (r *CategoryRepo) LockCategory(ctx context.Context, catID uuid.UUID) (*CategoryRow, error) {
	var cat CategoryRow
	var createdAt, updatedAt sql.NullTime
	var pillarID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, created_by, updated_by, name, COALESCE(description, ''), weight, employee_id, pillar_id
		 FROM goal_categories WHERE id = $1 FOR UPDATE`,
		catID,
	).Scan(&cat.ID, &createdAt, &updatedAt, &cat.CreatedBy, &cat.UpdatedBy,
		&cat.Name, &cat.Description, &cat.Weight, &cat.EmployeeID, &pillarID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrCategoryNotFound
		}
		return nil, err
	}
	if createdAt.Valid {
		cat.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		cat.UpdatedAt = updatedAt.Time
	}
	if pillarID.Valid {
		u := uuid.MustParse(pillarID.String)
		cat.PillarID = &u
	}
	return &cat, nil
}
