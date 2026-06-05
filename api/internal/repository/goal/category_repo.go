// Package goal provides the repository layer for goal-related entities.
package goal

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/goalcategory"
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
}

// rowToCategoryRow converts an ent GoalCategory to a CategoryRow.
func rowToCategoryRow(c *internal.GoalCategory) *CategoryRow {
	if c == nil {
		return nil
	}
	return &CategoryRow{
		ID:          c.ID,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		CreatedBy:   c.CreatedBy,
		UpdatedBy:   c.UpdatedBy,
		Name:        c.Name,
		Description: c.Description,
		Weight:      c.Weight,
		EmployeeID:  c.EmployeeID,
	}
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
	cats, err := r.client.GoalCategory.Query().
		Where(goalcategory.EmployeeID(empID)).
		Order(internal.Asc(goalcategory.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]*CategoryRow, len(cats))
	for i, c := range cats {
		rows[i] = rowToCategoryRow(c)
	}
	return rows, nil
}

// CreateCategory inserts a new category.
func (r *CategoryRepo) CreateCategory(ctx context.Context, empID uuid.UUID, name, description string, weight float64) (*CategoryRow, error) {
	cat, err := r.client.GoalCategory.Create().
		SetEmployeeID(empID).
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
	return rowToCategoryRow(cat), nil
}

// UpdateCategory updates an existing category's name, description, and weight.
func (r *CategoryRepo) UpdateCategory(ctx context.Context, catID uuid.UUID, name, description string, weight float64) (*CategoryRow, error) {
	cat, err := r.client.GoalCategory.UpdateOneID(catID).
		SetName(name).
		SetDescription(description).
		SetWeight(weight).
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
	return rowToCategoryRow(cat), nil
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
	cat, err := r.client.GoalCategory.Query().
		Where(goalcategory.ID(catID)).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.ErrCategoryNotFound
		}
		return nil, err
	}
	return rowToCategoryRow(cat), nil
}

// LockCategory acquires a SELECT FOR UPDATE lock on a category row.
// This serialises weight-sum calculations for that category.
// Uses raw SQL because Ent's query builder does not expose ForUpdate().
func (r *CategoryRepo) LockCategory(ctx context.Context, catID uuid.UUID) (*CategoryRow, error) {
	// Use a raw query with FOR UPDATE via the sql.DB connection
	row := r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, created_by, updated_by, name, COALESCE(description, ''), weight, employee_id
		 FROM goal_categories WHERE id = $1 FOR UPDATE`,
		catID,
	)

	var cat CategoryRow
	var createdAt, updatedAt sql.NullTime
	err := row.Scan(&cat.ID, &createdAt, &updatedAt, &cat.CreatedBy, &cat.UpdatedBy,
		&cat.Name, &cat.Description, &cat.Weight, &cat.EmployeeID)
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
	return &cat, nil
}
