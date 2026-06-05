package goal

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// WeightQueries holds aggregate weight sum queries.
type WeightQueries struct {
	db *sql.DB
}

// NewWeightQueries creates a new WeightQueries.
func NewWeightQueries(db *sql.DB) *WeightQueries {
	return &WeightQueries{db: db}
}

// SumGoalWeightsByCategoryID returns the sum of all goal weights for a category.
// Returns 0.0 for an empty category (no goals).
func (q *WeightQueries) SumGoalWeightsByCategoryID(ctx context.Context, catID uuid.UUID) (float64, error) {
	var sum sql.NullFloat64
	err := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(weight), 0) FROM goals WHERE category_id = $1`,
		catID,
	).Scan(&sum)
	if err != nil {
		return 0, err
	}
	if sum.Valid {
		return sum.Float64, nil
	}
	return 0, nil
}

// SumCategoryWeightsByEmployee returns the sum of all category weights for an employee.
// Returns 0.0 for an employee with no categories.
func (q *WeightQueries) SumCategoryWeightsByEmployee(ctx context.Context, empID uuid.UUID) (float64, error) {
	var sum sql.NullFloat64
	err := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(weight), 0) FROM goal_categories WHERE employee_id = $1`,
		empID,
	).Scan(&sum)
	if err != nil {
		return 0, err
	}
	if sum.Valid {
		return sum.Float64, nil
	}
	return 0, nil
}
