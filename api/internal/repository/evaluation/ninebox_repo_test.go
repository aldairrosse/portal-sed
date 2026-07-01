package evaluation_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
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

func TestNineBoxRepo_GetEmployeesByIDs_Success(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewNineBoxRepo(nil, db)

	id1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	id2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	profileID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	mock.ExpectQuery("SELECT id, first_name, last_name, profile_id FROM employees WHERE id IN \\(\\$1,\\$2\\)").
		WithArgs(id1, id2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "profile_id"}).
			AddRow(id1, "María", "García", profileID).
			AddRow(id2, "Carlos", "López", uuid.Nil))

	infoMap, err := r.GetEmployeesByIDs(context.Background(), []uuid.UUID{id1, id2})
	require.NoError(t, err)
	require.Len(t, infoMap, 2)
	assert.Equal(t, "María", infoMap[id1].FirstName)
	assert.Equal(t, "García", infoMap[id1].LastName)
	assert.Equal(t, profileID, infoMap[id1].ProfileID)
	assert.Equal(t, "Carlos", infoMap[id2].FirstName)
	assert.Equal(t, "López", infoMap[id2].LastName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNineBoxRepo_GetEmployeesByIDs_EmptyIDs(t *testing.T) {
	t.Parallel()

	db, _ := newMockDB(t)
	r := repo.NewNineBoxRepo(nil, db)

	infoMap, err := r.GetEmployeesByIDs(context.Background(), []uuid.UUID{})
	require.NoError(t, err)
	assert.Empty(t, infoMap)
}

func TestNineBoxRepo_GetManagerMapping_Success(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewNineBoxRepo(nil, db)

	empID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	managerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	rootID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	mock.ExpectQuery("SELECT id, manager_id FROM employees WHERE id IN \\(\\$1,\\$2\\) AND manager_id IS NOT NULL").
		WithArgs(empID, rootID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manager_id"}).
			AddRow(empID, managerID))

	mapping, err := r.GetManagerMapping(context.Background(), []uuid.UUID{empID, rootID})
	require.NoError(t, err)
	require.Len(t, mapping, 1)
	assert.Equal(t, managerID, mapping[empID])
	_, ok := mapping[rootID]
	assert.False(t, ok, "root employee with NULL manager_id should be skipped")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNineBoxRepo_GetManagerMapping_EmptyIDs(t *testing.T) {
	t.Parallel()

	db, _ := newMockDB(t)
	r := repo.NewNineBoxRepo(nil, db)

	mapping, err := r.GetManagerMapping(context.Background(), []uuid.UUID{})
	require.NoError(t, err)
	assert.Empty(t, mapping)
}

func TestNineBoxRepo_GetMatrixEntriesByQuadrant_Success(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewNineBoxRepo(nil, db)

	matrixID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	entryID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	evaluateeID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	now := time.Now()

	mock.ExpectQuery("SELECT id, created_at, updated_at, matrix_id, evaluatee_id, performance_tier, potential_tier, quadrant, comments FROM nine_box_entries WHERE matrix_id = \\$1 AND quadrant = \\$2").
		WithArgs(matrixID, 5).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "matrix_id", "evaluatee_id",
			"performance_tier", "potential_tier", "quadrant", "comments",
		}).AddRow(entryID, now, now, matrixID, evaluateeID, 2, 2, 5, ""))

	entries, err := r.GetMatrixEntriesByQuadrant(context.Background(), matrixID, 5)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, entryID, entries[0].ID)
	assert.Equal(t, 5, entries[0].Quadrant)

	assert.NoError(t, mock.ExpectationsWereMet())
}
