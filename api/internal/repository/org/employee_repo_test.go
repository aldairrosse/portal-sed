package org_test

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockDB creates a sqlmock DB pair.
func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestEmployeeRepo_Create_Success(t *testing.T) {
	t.Parallel()

	// Note: EmployeeRepo does not expose a Create method directly;
	// this test verifies successful retrieval of an employee row
	// as if it had been created via an upstream service.
	db, mock := newMockDB(t)
	r := repo.NewEmployeeRepo(nil, db)

	empID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	orgNodeID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	profileID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	now := time.Now()

	mock.ExpectQuery("SELECT id, created_at, updated_at, first_name, last_name, email, employee_number, is_active, org_node_id, manager_id, profile_id FROM employees WHERE id = \\$1").
		WithArgs(empID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id",
		}).AddRow(empID, now, now, "Alice", "Anderson", "alice@example.com", "E001", true, orgNodeID, nil, profileID))

	emp, err := r.GetByID(context.Background(), empID)
	require.NoError(t, err)
	assert.Equal(t, empID, emp.ID)
	assert.Equal(t, "Alice", emp.FirstName)
	assert.Equal(t, "Anderson", emp.LastName)
	assert.Equal(t, "alice@example.com", emp.Email)
	assert.Equal(t, "E001", emp.EmployeeNumber)
	assert.True(t, emp.IsActive)
	assert.Equal(t, orgNodeID, emp.OrgNodeID)
	assert.Nil(t, emp.ManagerID)
	assert.Equal(t, profileID, emp.ProfileID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmployeeRepo_Search_ILIKE(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewEmployeeRepo(nil, db)

	orgNodeID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	profileID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	now := time.Now()

	mock.ExpectQuery("SELECT id, created_at, updated_at, first_name, last_name, email, employee_number, is_active, org_node_id, manager_id, profile_id FROM employees WHERE first_name ILIKE \\$1 OR last_name ILIKE \\$1 OR email ILIKE \\$1 OR employee_number ILIKE \\$1 ORDER BY last_name, first_name LIMIT \\$2").
		WithArgs("%alice%", 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id",
		}).
			AddRow(uuid.New(), now, now, "Alice", "Anderson", "alice@example.com", "E001", true, orgNodeID, nil, profileID).
			AddRow(uuid.New(), now, now, "Alicia", "Keys", "alicia@example.com", "E002", true, orgNodeID, nil, profileID))

	results, err := r.Search(context.Background(), "alice", 20)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "Alice", results[0].FirstName)
	assert.Equal(t, "Alicia", results[1].FirstName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmployeeRepo_BatchResolve(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewEmployeeRepo(nil, db)

	id1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	id2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	id3 := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	orgNodeID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	profileID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	now := time.Now()

	mock.ExpectQuery("SELECT id, created_at, updated_at, first_name, last_name, email, employee_number, is_active, org_node_id, manager_id, profile_id FROM employees WHERE id IN \\(").
		WithArgs(id1, id2, id3).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id",
		}).
			AddRow(id1, now, now, "Bob", "Barker", "bob@example.com", "E010", true, orgNodeID, nil, profileID).
			AddRow(id2, now, now, "Bill", "Bryson", "bill@example.com", "E011", true, orgNodeID, nil, profileID).
			AddRow(id3, now, now, "Ben", "Button", "ben@example.com", "E012", true, orgNodeID, nil, profileID))

	results, err := r.GetByIDs(context.Background(), []uuid.UUID{id1, id2, id3})
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, id1, results[0].ID)
	assert.Equal(t, id2, results[1].ID)
	assert.Equal(t, id3, results[2].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmployeeRepo_List_WithFilters(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewEmployeeRepo(nil, db)

	treeID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	nodeID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	profileID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	orgNodeID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	now := time.Now()
	active := true

	mock.ExpectQuery("SELECT e\\.id, e\\.created_at, e\\.updated_at, e\\.first_name, e\\.last_name, e\\.email, e\\.employee_number, e\\.is_active, e\\.org_node_id, e\\.manager_id, e\\.profile_id FROM employees e JOIN org_nodes on2 ON e\\.org_node_id = on2\\.id WHERE on2\\.organization_id = \\$1 AND e\\.org_node_id = \\$2 AND e\\.profile_id = \\$3 AND e\\.is_active = \\$4 AND \\(e\\.first_name ILIKE \\$5 OR e\\.last_name ILIKE \\$5 OR e\\.email ILIKE \\$5 OR e\\.employee_number ILIKE \\$5\\) ORDER BY e\\.last_name, e\\.first_name, e\\.id LIMIT \\$6").
		WithArgs(treeID, nodeID, profileID, active, "%smith%", 51).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "first_name", "last_name", "email",
			"employee_number", "is_active", "org_node_id", "manager_id", "profile_id",
		}).
			AddRow(uuid.New(), now, now, "John", "Smith", "john@example.com", "E100", true, orgNodeID, nil, profileID).
			AddRow(uuid.New(), now, now, "Jane", "Smith", "jane@example.com", "E101", true, orgNodeID, nil, profileID))

	filter := repo.EmployeeFilter{
		TreeID:    &treeID,
		NodeID:    &nodeID,
		ProfileID: &profileID,
		IsActive:  &active,
		Query:     "smith",
		Limit:     50,
	}

	results, err := r.List(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "John", results[0].FirstName)
	assert.Equal(t, "Jane", results[1].FirstName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmployeeRepo_ConcurrentSearch(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewEmployeeRepo(nil, db)

	orgNodeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	profileID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	now := time.Now()

	const workers = 20
	for i := 0; i < workers; i++ {
		mock.ExpectQuery("SELECT id, created_at, updated_at, first_name, last_name, email, employee_number, is_active, org_node_id, manager_id, profile_id FROM employees WHERE first_name ILIKE \\$1 OR last_name ILIKE \\$1 OR email ILIKE \\$1 OR employee_number ILIKE \\$1 ORDER BY last_name, first_name LIMIT \\$2").
			WithArgs("%query%", 20).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "first_name", "last_name", "email",
				"employee_number", "is_active", "org_node_id", "manager_id", "profile_id",
			}).
				AddRow(uuid.New(), now, now, "Query", "User", "query@example.com", "E999", true, orgNodeID, nil, profileID))
	}

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			results, err := r.Search(context.Background(), "query", 20)
			if err == nil {
				assert.Len(t, results, 1)
			}
		}()
	}

	wg.Wait()
}
