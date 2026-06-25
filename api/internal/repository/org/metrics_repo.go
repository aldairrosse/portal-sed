package org

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
)

// GoalRow is a read model for goals with their employee association.
type GoalRow struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	TargetValue  float64   `json:"target_value"`
	CurrentValue float64   `json:"current_value"`
	State        string    `json:"state"`
	EmployeeID   uuid.UUID `json:"employee_id"`
}

// RHRatingRow is a read model for RH evaluation ratings.
type RHRatingRow struct {
	RHRating   float64   `json:"rh_rating"`
	EmployeeID uuid.UUID `json:"employee_id"`
}

// MetricsRepo provides database queries for area metrics aggregation.
type MetricsRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewMetricsRepo creates a new MetricsRepo.
func NewMetricsRepo(client *internal.Client, db *sql.DB) *MetricsRepo {
	return &MetricsRepo{client: client, db: db}
}

// GetDirectEmployees returns active employees directly assigned to the given org node.
// Results are ordered by last_name, first_name.
func (r *MetricsRepo) GetDirectEmployees(ctx context.Context, nodeID uuid.UUID) ([]*EmployeeRow, error) {
	return scanEmployeeRows(r.db, ctx,
		`SELECT id, created_at, updated_at, first_name, last_name, email,
		        employee_number, is_active, org_node_id, manager_id, profile_id
		 FROM employees
		 WHERE org_node_id = $1
		   AND is_active = true
		 ORDER BY last_name, first_name`, nodeID)
}

// GetGoalsByEmployees returns goals for the given employees.
// Joins through goal_categories to link goals to employees.
// Excludes goals where target_value = 0 to avoid division by zero.
func (r *MetricsRepo) GetGoalsByEmployees(ctx context.Context, employeeIDs []uuid.UUID) ([]*GoalRow, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(employeeIDs))
	args := make([]interface{}, len(employeeIDs))
	for i, id := range employeeIDs {
		placeholders[i] = "$" + itoa(i+1)
		args[i] = id
	}

	query := `SELECT g.id, g.name, g.target_value, g.current_value, g.state,
	                  gc.employee_id
	           FROM goals g
	           JOIN goal_categories gc ON gc.id = g.category_id
	           WHERE gc.employee_id IN (` + strings.Join(placeholders, ",") + `)
	             AND g.target_value > 0`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*GoalRow
	for rows.Next() {
		g := &GoalRow{}
		if err := rows.Scan(&g.ID, &g.Name, &g.TargetValue, &g.CurrentValue, &g.State, &g.EmployeeID); err != nil {
			return nil, err
		}
		results = append(results, g)
	}
	return results, rows.Err()
}

// GetRHEvaluationsByEmployees returns RH evaluation ratings for the given
// employees and cycle. Only includes competencies with a non-null rh_rating.
func (r *MetricsRepo) GetRHEvaluationsByEmployees(ctx context.Context, employeeIDs []uuid.UUID, cycleID uuid.UUID) ([]*RHRatingRow, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(employeeIDs))
	args := make([]interface{}, len(employeeIDs)+1)
	for i, id := range employeeIDs {
		placeholders[i] = "$" + itoa(i+1)
		args[i] = id
	}
	args[len(employeeIDs)] = cycleID

	cycleParamIdx := len(employeeIDs) + 1

	query := `SELECT ec.rh_rating, e.employee_id
	           FROM evaluation_competencies ec
	           JOIN evaluations e ON e.id = ec.evaluation_id
	           WHERE e.employee_id IN (` + strings.Join(placeholders, ",") + `)
	             AND e.cycle_id = $` + itoa(cycleParamIdx) + `
	             AND ec.rh_rating IS NOT NULL`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*RHRatingRow
	for rows.Next() {
		rr := &RHRatingRow{}
		if err := rows.Scan(&rr.RHRating, &rr.EmployeeID); err != nil {
			return nil, err
		}
		results = append(results, rr)
	}
	return results, rows.Err()
}
