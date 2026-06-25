package org_test

import (
	"context"
	"sync"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsRepo_GetDirectEmployees_Success(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewMetricsRepo(nil, db)

	nodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	empID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	profileID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	now := time.Now()

	mock.ExpectQuery("SELECT id, created_at, updated_at, first_name, last_name, email, employee_number, is_active, org_node_id, manager_id, profile_id FROM employees WHERE org_node_id = \\$1 AND is_active = true ORDER BY last_name, first_name").
		WithArgs(nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id",
		}).AddRow(empID, now, now, "Alice", "Smith", "alice@example.com", "E001", true, nodeID, nil, profileID))

	employees, err := r.GetDirectEmployees(context.Background(), nodeID)
	require.NoError(t, err)
	require.Len(t, employees, 1)
	assert.Equal(t, empID, employees[0].ID)
	assert.Equal(t, "Alice", employees[0].FirstName)
	assert.Equal(t, "Smith", employees[0].LastName)
	assert.Equal(t, profileID, employees[0].ProfileID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_GetDirectEmployees_NoEmployees(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewMetricsRepo(nil, db)

	nodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	mock.ExpectQuery("SELECT id, created_at, updated_at, first_name, last_name, email, employee_number, is_active, org_node_id, manager_id, profile_id FROM employees WHERE org_node_id = \\$1 AND is_active = true ORDER BY last_name, first_name").
		WithArgs(nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id",
		}))

	employees, err := r.GetDirectEmployees(context.Background(), nodeID)
	require.NoError(t, err)
	assert.Empty(t, employees)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_GetGoalsByEmployees_Success(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewMetricsRepo(nil, db)

	empID1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	empID2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	goalID1 := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	goalID2 := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	mock.ExpectQuery("SELECT g.id, g.name, g.target_value, g.current_value, g.state, gc.employee_id FROM goals g JOIN goal_categories gc ON gc.id = g.category_id WHERE gc.employee_id IN \\(\\$1,\\$2\\) AND g.target_value > 0").
		WithArgs(empID1, empID2).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "target_value", "current_value", "state", "employee_id",
		}).
			AddRow(goalID1, "Goal A", 100.0, 75.0, "in_progress", empID1).
			AddRow(goalID2, "Goal B", 50.0, 50.0, "completed", empID1))

	goals, err := r.GetGoalsByEmployees(context.Background(), []uuid.UUID{empID1, empID2})
	require.NoError(t, err)
	require.Len(t, goals, 2)
	assert.Equal(t, goalID1, goals[0].ID)
	assert.Equal(t, "Goal A", goals[0].Name)
	assert.Equal(t, 100.0, goals[0].TargetValue)
	assert.Equal(t, 75.0, goals[0].CurrentValue)
	assert.Equal(t, "in_progress", goals[0].State)
	assert.Equal(t, empID1, goals[0].EmployeeID)
	assert.Equal(t, goalID2, goals[1].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_GetGoalsByEmployees_EmptyIDs(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewMetricsRepo(nil, db)

	goals, err := r.GetGoalsByEmployees(context.Background(), []uuid.UUID{})
	require.NoError(t, err)
	assert.Nil(t, goals)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_GetRHEvaluationsByEmployees_Success(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewMetricsRepo(nil, db)

	empID1 := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	empID2 := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	cycleID := uuid.MustParse("77777777-7777-7777-7777-777777777777")

	mock.ExpectQuery("SELECT ec.rh_rating, e.employee_id FROM evaluation_competencies ec JOIN evaluations e ON e.id = ec.evaluation_id WHERE e.employee_id IN \\(\\$1,\\$2\\) AND e.cycle_id = \\$3 AND ec.rh_rating IS NOT NULL").
		WithArgs(empID1, empID2, cycleID).
		WillReturnRows(sqlmock.NewRows([]string{
			"rh_rating", "employee_id",
		}).
			AddRow(4.5, empID1).
			AddRow(3.0, empID2))

	ratings, err := r.GetRHEvaluationsByEmployees(context.Background(), []uuid.UUID{empID1, empID2}, cycleID)
	require.NoError(t, err)
	require.Len(t, ratings, 2)
	assert.Equal(t, 4.5, ratings[0].RHRating)
	assert.Equal(t, empID1, ratings[0].EmployeeID)
	assert.Equal(t, 3.0, ratings[1].RHRating)
	assert.Equal(t, empID2, ratings[1].EmployeeID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_GetRHEvaluationsByEmployees_EmptyIDs(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewMetricsRepo(nil, db)

	cycleID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	ratings, err := r.GetRHEvaluationsByEmployees(context.Background(), []uuid.UUID{}, cycleID)
	require.NoError(t, err)
	assert.Nil(t, ratings)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_GetRHEvaluationsByEmployees_NoRatings(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewMetricsRepo(nil, db)

	empID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	cycleID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	mock.ExpectQuery("SELECT ec.rh_rating, e.employee_id FROM evaluation_competencies ec JOIN evaluations e ON e.id = ec.evaluation_id WHERE e.employee_id IN \\(\\$1\\) AND e.cycle_id = \\$2 AND ec.rh_rating IS NOT NULL").
		WithArgs(empID, cycleID).
		WillReturnRows(sqlmock.NewRows([]string{
			"rh_rating", "employee_id",
		}))

	ratings, err := r.GetRHEvaluationsByEmployees(context.Background(), []uuid.UUID{empID}, cycleID)
	require.NoError(t, err)
	assert.Empty(t, ratings)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_GetGoalsByEmployees_NoGoals(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewMetricsRepo(nil, db)

	empID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	mock.ExpectQuery("SELECT g.id, g.name, g.target_value, g.current_value, g.state, gc.employee_id FROM goals g JOIN goal_categories gc ON gc.id = g.category_id WHERE gc.employee_id IN \\(\\$1\\) AND g.target_value > 0").
		WithArgs(empID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "target_value", "current_value", "state", "employee_id",
		}))

	goals, err := r.GetGoalsByEmployees(context.Background(), []uuid.UUID{empID})
	require.NoError(t, err)
	assert.Empty(t, goals)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_ConcurrentQueries(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewMetricsRepo(nil, db)

	nodeID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	empID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	profileID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	now := time.Now()

	const workers = 10
	for i := 0; i < workers; i++ {
		mock.ExpectQuery("SELECT id, created_at, updated_at, first_name, last_name, email, employee_number, is_active, org_node_id, manager_id, profile_id FROM employees WHERE org_node_id = \\$1 AND is_active = true ORDER BY last_name, first_name").
			WithArgs(nodeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "first_name", "last_name", "email",
				"employee_number", "is_active", "org_node_id", "manager_id", "profile_id",
			}).AddRow(empID, now, now, "Concurrent", "User", "cu@example.com", "E999", true, nodeID, nil, profileID))
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			employees, err := r.GetDirectEmployees(context.Background(), nodeID)
			if err == nil {
				assert.Len(t, employees, 1)
			}
		}()
	}
	wg.Wait()
}
