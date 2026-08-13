package org_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/org"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMetricsService(db *sql.DB, mock sqlmock.Sqlmock) svc.MetricsService {
	metricsRepo := repo.NewMetricsRepo(nil, db)
	nodeRepo := repo.NewOrgNodeRepo(nil, db)
	return svc.NewMetricsService(metricsRepo, nodeRepo, nil)
}

func newOrgNodeRow(id, orgID uuid.UUID, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "name", "type", "code",
		"organization_id", "parent_id", "path", "version", "head_employee_id",
	}).AddRow(id, now, now, "Engineering", "corporate", "ENG", orgID, nil, "1", 1, nil)
}

func TestMetricsService_GetAreaMetrics_Success(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	service := newMetricsService(db, mock)

	nodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	orgID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	empID1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	empID2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	profileID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	goalID1 := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	goalID2 := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	cycleID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	now := time.Now()

	// Expect node existence check
	mock.ExpectQuery("SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id, COALESCE\\(path::text, ''\\) as path, COALESCE\\(version, 0\\), head_employee_id FROM org_nodes WHERE id = \\$1").
		WithArgs(nodeID).
		WillReturnRows(newOrgNodeRow(nodeID, orgID, now))

	// Expect GetDirectEmployees
	mock.ExpectQuery("SELECT e\\.id, e\\.created_at, e\\.updated_at, e\\.first_name, e\\.last_name, e\\.email, e\\.employee_number, e\\.is_active, e\\.org_node_id, e\\.manager_id, e\\.profile_id, COALESCE\\(ep\\.name, ''\\) as profile_name, COALESCE\\(ep\\.description, ''\\) as profile_description, e\\.job_title FROM employees e LEFT JOIN evaluation_profiles ep ON e\\.profile_id = ep\\.id WHERE e\\.org_node_id = \\$1 AND e\\.is_active = true ORDER BY e\\.last_name, e\\.first_name").
		WithArgs(nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id", "profile_name", "profile_description", "job_title",
		}).
			AddRow(empID1, now, now, "Alice", "Smith", "alice@example.com", "E001", true, nodeID, nil, profileID, "", "", "").
			AddRow(empID2, now, now, "Bob", "Jones", "bob@example.com", "E002", true, nodeID, nil, profileID, "", "", ""))

	// Expect GetGoalsByEmployees
	mock.ExpectQuery("SELECT g.id, g.name, g.target_value, g.current_value, g.state, gc.employee_id FROM goals g JOIN goal_categories gc ON gc.id = g.category_id WHERE gc.employee_id IN \\(\\$1,\\$2\\) AND g.target_value > 0").
		WithArgs(empID1, empID2).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "target_value", "current_value", "state", "employee_id",
		}).
			AddRow(goalID1, "Goal A", 100.0, 75.0, "in_progress", empID1).
			AddRow(goalID2, "Goal B", 50.0, 50.0, "completed", empID2))

	// Expect GetRHEvaluationsByEmployees
	mock.ExpectQuery("SELECT ec.rh_rating, e.employee_id FROM evaluation_competencies ec JOIN evaluations e ON e.id = ec.evaluation_id WHERE e.employee_id IN \\(\\$1,\\$2\\) AND e.cycle_id = \\$3 AND ec.rh_rating IS NOT NULL").
		WithArgs(empID1, empID2, cycleID).
		WillReturnRows(sqlmock.NewRows([]string{
			"rh_rating", "employee_id",
		}).
			AddRow(4.5, empID1).
			AddRow(3.5, empID2))

	resp, err := service.GetAreaMetrics(context.Background(), nodeID.String(), cycleID.String())
	require.NoError(t, err)
	assert.Equal(t, nodeID.String(), resp.NodeID)
	assert.Equal(t, 2, resp.EmployeeCount)
	assert.Equal(t, 2, resp.EmployeesWithGoals)

	// Expected avg progress: ((75/100)*100 + (50/50)*100) / 2 = (75 + 100) / 2 = 87.5
	require.NotNil(t, resp.AvgProgress)
	assert.Equal(t, 87.5, *resp.AvgProgress)

	assert.Equal(t, 1, resp.CompletedGoals) // Goal B: 50 >= 50
	assert.Equal(t, 1, resp.PendingGoals)   // Goal A: 75 < 100

	// Expected avg rating: (4.5 + 3.5) / 2 = 4.0
	require.NotNil(t, resp.AvgRating)
	assert.Equal(t, 4.0, *resp.AvgRating)

	assert.Equal(t, 2, resp.RatingsCount)
	require.Len(t, resp.Employees, 2)
	assert.Equal(t, empID1.String(), resp.Employees[0].ID)
	assert.Equal(t, "Alice", resp.Employees[0].FirstName)
	assert.Equal(t, profileID.String(), resp.Employees[0].ProfileID)
	assert.Equal(t, empID2.String(), resp.Employees[1].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsService_GetAreaMetrics_NoEmployees(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	service := newMetricsService(db, mock)

	nodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	orgID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	now := time.Now()

	// Expect node existence check
	mock.ExpectQuery("SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id, COALESCE\\(path::text, ''\\) as path, COALESCE\\(version, 0\\), head_employee_id FROM org_nodes WHERE id = \\$1").
		WithArgs(nodeID).
		WillReturnRows(newOrgNodeRow(nodeID, orgID, now))

	// Expect empty employees
	mock.ExpectQuery("SELECT e\\.id, e\\.created_at, e\\.updated_at, e\\.first_name, e\\.last_name, e\\.email, e\\.employee_number, e\\.is_active, e\\.org_node_id, e\\.manager_id, e\\.profile_id, COALESCE\\(ep\\.name, ''\\) as profile_name, COALESCE\\(ep\\.description, ''\\) as profile_description, e\\.job_title FROM employees e LEFT JOIN evaluation_profiles ep ON e\\.profile_id = ep\\.id WHERE e\\.org_node_id = \\$1 AND e\\.is_active = true ORDER BY e\\.last_name, e\\.first_name").
		WithArgs(nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id", "profile_name", "profile_description", "job_title",
		}))

	resp, err := service.GetAreaMetrics(context.Background(), nodeID.String(), "")
	require.NoError(t, err)
	assert.Equal(t, nodeID.String(), resp.NodeID)
	assert.Equal(t, 0, resp.EmployeeCount)
	assert.Equal(t, 0, resp.EmployeesWithGoals)
	assert.Nil(t, resp.AvgProgress)
	assert.Equal(t, 0, resp.CompletedGoals)
	assert.Equal(t, 0, resp.PendingGoals)
	assert.Nil(t, resp.AvgRating)
	assert.Equal(t, 0, resp.RatingsCount)
	assert.Empty(t, resp.Employees)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsService_GetAreaMetrics_NoGoals(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	service := newMetricsService(db, mock)

	nodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	orgID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	empID1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	profileID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	now := time.Now()

	// Expect node existence check
	mock.ExpectQuery("SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id, COALESCE\\(path::text, ''\\) as path, COALESCE\\(version, 0\\), head_employee_id FROM org_nodes WHERE id = \\$1").
		WithArgs(nodeID).
		WillReturnRows(newOrgNodeRow(nodeID, orgID, now))

	// Expect GetDirectEmployees
	mock.ExpectQuery("SELECT e\\.id, e\\.created_at, e\\.updated_at, e\\.first_name, e\\.last_name, e\\.email, e\\.employee_number, e\\.is_active, e\\.org_node_id, e\\.manager_id, e\\.profile_id, COALESCE\\(ep\\.name, ''\\) as profile_name, COALESCE\\(ep\\.description, ''\\) as profile_description, e\\.job_title FROM employees e LEFT JOIN evaluation_profiles ep ON e\\.profile_id = ep\\.id WHERE e\\.org_node_id = \\$1 AND e\\.is_active = true ORDER BY e\\.last_name, e\\.first_name").
		WithArgs(nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id", "profile_name", "profile_description", "job_title",
		}).AddRow(empID1, now, now, "Alice", "Smith", "alice@example.com", "E001", true, nodeID, nil, profileID, "", "", ""))

	// Expect empty goals
	mock.ExpectQuery("SELECT g.id, g.name, g.target_value, g.current_value, g.state, gc.employee_id FROM goals g JOIN goal_categories gc ON gc.id = g.category_id WHERE gc.employee_id IN \\(\\$1\\) AND g.target_value > 0").
		WithArgs(empID1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "target_value", "current_value", "state", "employee_id",
		}))

	// No cycleID provided, so no RH evaluations query

	resp, err := service.GetAreaMetrics(context.Background(), nodeID.String(), "")
	require.NoError(t, err)
	assert.Equal(t, 1, resp.EmployeeCount)
	assert.Equal(t, 0, resp.EmployeesWithGoals)
	assert.Nil(t, resp.AvgProgress)
	assert.Equal(t, 0, resp.CompletedGoals)
	assert.Equal(t, 0, resp.PendingGoals)
	assert.Nil(t, resp.AvgRating)
	assert.Equal(t, 0, resp.RatingsCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsService_GetAreaMetrics_NoRatings(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	service := newMetricsService(db, mock)

	nodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	orgID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	empID1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	profileID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	goalID1 := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	cycleID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	now := time.Now()

	// Expect node existence check
	mock.ExpectQuery("SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id, COALESCE\\(path::text, ''\\) as path, COALESCE\\(version, 0\\), head_employee_id FROM org_nodes WHERE id = \\$1").
		WithArgs(nodeID).
		WillReturnRows(newOrgNodeRow(nodeID, orgID, now))

	// Expect GetDirectEmployees
	mock.ExpectQuery("SELECT e\\.id, e\\.created_at, e\\.updated_at, e\\.first_name, e\\.last_name, e\\.email, e\\.employee_number, e\\.is_active, e\\.org_node_id, e\\.manager_id, e\\.profile_id, COALESCE\\(ep\\.name, ''\\) as profile_name, COALESCE\\(ep\\.description, ''\\) as profile_description, e\\.job_title FROM employees e LEFT JOIN evaluation_profiles ep ON e\\.profile_id = ep\\.id WHERE e\\.org_node_id = \\$1 AND e\\.is_active = true ORDER BY e\\.last_name, e\\.first_name").
		WithArgs(nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id", "profile_name", "profile_description", "job_title",
		}).AddRow(empID1, now, now, "Alice", "Smith", "alice@example.com", "E001", true, nodeID, nil, profileID, "", "", ""))

	// Expect GetGoalsByEmployees
	mock.ExpectQuery("SELECT g.id, g.name, g.target_value, g.current_value, g.state, gc.employee_id FROM goals g JOIN goal_categories gc ON gc.id = g.category_id WHERE gc.employee_id IN \\(\\$1\\) AND g.target_value > 0").
		WithArgs(empID1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "target_value", "current_value", "state", "employee_id",
		}).AddRow(goalID1, "Goal A", 100.0, 50.0, "in_progress", empID1))

	// Expect empty ratings
	mock.ExpectQuery("SELECT ec.rh_rating, e.employee_id FROM evaluation_competencies ec JOIN evaluations e ON e.id = ec.evaluation_id WHERE e.employee_id IN \\(\\$1\\) AND e.cycle_id = \\$2 AND ec.rh_rating IS NOT NULL").
		WithArgs(empID1, cycleID).
		WillReturnRows(sqlmock.NewRows([]string{
			"rh_rating", "employee_id",
		}))

	resp, err := service.GetAreaMetrics(context.Background(), nodeID.String(), cycleID.String())
	require.NoError(t, err)
	assert.Equal(t, 1, resp.EmployeeCount)

	// Goals exist so progress should be computed
	require.NotNil(t, resp.AvgProgress)
	assert.Equal(t, 50.0, *resp.AvgProgress)
	assert.Equal(t, 0, resp.CompletedGoals)
	assert.Equal(t, 1, resp.PendingGoals)

	// No ratings so avgRating is nil
	assert.Nil(t, resp.AvgRating)
	assert.Equal(t, 0, resp.RatingsCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsService_GetAreaMetrics_NodeNotFound(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	service := newMetricsService(db, mock)

	nodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	// Expect node existence check - no rows
	mock.ExpectQuery("SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id, COALESCE\\(path::text, ''\\) as path, COALESCE\\(version, 0\\), head_employee_id FROM org_nodes WHERE id = \\$1").
		WithArgs(nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "type", "code",
			"organization_id", "parent_id", "path", "version",
		}))

	_, err := service.GetAreaMetrics(context.Background(), nodeID.String(), "")
	require.Error(t, err)

	var de *errors.DomainError
	require.True(t, errors.AsDomainError(err, &de))
	assert.Equal(t, "NODE_NOT_FOUND", string(de.Code))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsService_GetAreaMetrics_InvalidNodeID(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	service := newMetricsService(db, mock)

	_, err := service.GetAreaMetrics(context.Background(), "not-a-uuid", "")
	require.Error(t, err)

	var de *errors.DomainError
	require.True(t, errors.AsDomainError(err, &de))
	assert.Equal(t, "INVALID_REQUEST", string(de.Code))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsService_GetAreaMetrics_InvalidCycleID(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	service := newMetricsService(db, mock)

	nodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	orgID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	empID1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	profileID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	now := time.Now()

	// Expect node existence check
	mock.ExpectQuery("SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id, COALESCE\\(path::text, ''\\) as path, COALESCE\\(version, 0\\), head_employee_id FROM org_nodes WHERE id = \\$1").
		WithArgs(nodeID).
		WillReturnRows(newOrgNodeRow(nodeID, orgID, now))

	// Expect GetDirectEmployees
	mock.ExpectQuery("SELECT e\\.id, e\\.created_at, e\\.updated_at, e\\.first_name, e\\.last_name, e\\.email, e\\.employee_number, e\\.is_active, e\\.org_node_id, e\\.manager_id, e\\.profile_id, COALESCE\\(ep\\.name, ''\\) as profile_name, COALESCE\\(ep\\.description, ''\\) as profile_description, e\\.job_title FROM employees e LEFT JOIN evaluation_profiles ep ON e\\.profile_id = ep\\.id WHERE e\\.org_node_id = \\$1 AND e\\.is_active = true ORDER BY e\\.last_name, e\\.first_name").
		WithArgs(nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id", "profile_name", "profile_description", "job_title",
		}).AddRow(empID1, now, now, "Alice", "Smith", "alice@example.com", "E001", true, nodeID, nil, profileID, "", "", ""))

	// Expect empty goals (no cycle filter needed, but we need goals query)
	mock.ExpectQuery("SELECT g.id, g.name, g.target_value, g.current_value, g.state, gc.employee_id FROM goals g JOIN goal_categories gc ON gc.id = g.category_id WHERE gc.employee_id IN \\(\\$1\\) AND g.target_value > 0").
		WithArgs(empID1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "target_value", "current_value", "state", "employee_id",
		}))

	// Provide an invalid cycleId - should cause error when parsing
	_, err := service.GetAreaMetrics(context.Background(), nodeID.String(), "not-a-uuid")
	require.Error(t, err)

	var de *errors.DomainError
	require.True(t, errors.AsDomainError(err, &de))
	assert.Equal(t, "INVALID_REQUEST", string(de.Code))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsService_GetAreaMetrics_NoCycleID(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	service := newMetricsService(db, mock)

	nodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	orgID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	empID1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	profileID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	goalID1 := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	now := time.Now()

	// Expect node existence check
	mock.ExpectQuery("SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id, COALESCE\\(path::text, ''\\) as path, COALESCE\\(version, 0\\), head_employee_id FROM org_nodes WHERE id = \\$1").
		WithArgs(nodeID).
		WillReturnRows(newOrgNodeRow(nodeID, orgID, now))

	// Expect GetDirectEmployees
	mock.ExpectQuery("SELECT e\\.id, e\\.created_at, e\\.updated_at, e\\.first_name, e\\.last_name, e\\.email, e\\.employee_number, e\\.is_active, e\\.org_node_id, e\\.manager_id, e\\.profile_id, COALESCE\\(ep\\.name, ''\\) as profile_name, COALESCE\\(ep\\.description, ''\\) as profile_description, e\\.job_title FROM employees e LEFT JOIN evaluation_profiles ep ON e\\.profile_id = ep\\.id WHERE e\\.org_node_id = \\$1 AND e\\.is_active = true ORDER BY e\\.last_name, e\\.first_name").
		WithArgs(nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id", "profile_name", "profile_description", "job_title",
		}).AddRow(empID1, now, now, "Alice", "Smith", "alice@example.com", "E001", true, nodeID, nil, profileID, "", "", ""))

	// Expect GetGoalsByEmployees
	mock.ExpectQuery("SELECT g.id, g.name, g.target_value, g.current_value, g.state, gc.employee_id FROM goals g JOIN goal_categories gc ON gc.id = g.category_id WHERE gc.employee_id IN \\(\\$1\\) AND g.target_value > 0").
		WithArgs(empID1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "target_value", "current_value", "state", "employee_id",
		}).AddRow(goalID1, "Goal A", 100.0, 80.0, "in_progress", empID1))

	// No cycleID, so no RH evaluations query
	resp, err := service.GetAreaMetrics(context.Background(), nodeID.String(), "")
	require.NoError(t, err)
	assert.Equal(t, 1, resp.EmployeeCount)

	// Goals should be computed even without cycleID
	require.NotNil(t, resp.AvgProgress)
	assert.Equal(t, 80.0, *resp.AvgProgress)

	// No cycleID -> no ratings queried
	assert.Nil(t, resp.AvgRating)
	assert.Equal(t, 0, resp.RatingsCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}
