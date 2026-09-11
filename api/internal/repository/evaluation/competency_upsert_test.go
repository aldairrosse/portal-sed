package evaluation_test

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	"github.com/stretchr/testify/require"
)

// lockRow stubs LockEvalForUpdate: state must not be "completada".
func stubLockRow(mock sqlmock.Sqlmock, evalID uuid.UUID) {
	now := time.Now()
	mock.ExpectQuery("SELECT e\\.id, e\\.created_at").
		WithArgs(evalID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "phase", "state",
			"self_evaluation_completed_at", "rh_evaluation_completed_at",
			"employee_id", "cycle_id", "version",
		}).AddRow(evalID, now, now, "avance", "en_progreso", nil, nil,
			uuid.New(), uuid.New(), 1))
}

// runSubmitEval executes SubmitEval for one competency and expects the
// upsert to target ON CONFLICT (evaluation_id, competency_id, source)
// with the given source value and a REAL profile_id. An exact WithArgs
// match on profileID means uuid.Nil fails the expectation.
func runSubmitEval(t *testing.T, source string, setSelf, setRH bool) {
	t.Helper()

	db, mock := newMockDB(t)
	r := repo.NewEvaluationRepo(nil, db)
	ctx := context.Background()

	evalID := uuid.New()
	compID := uuid.New()
	profileID := uuid.New()
	require.NotEqual(t, uuid.Nil, profileID)

	mock.ExpectBegin()
	stubLockRow(mock, evalID)
	mock.ExpectExec(`INSERT INTO evaluation_competencies.+ON CONFLICT \(evaluation_id, competency_id, source\)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			evalID, compID, 4, "good", profileID, source, 4).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE evaluations SET`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), evalID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE evaluations SET version = version \+ 1`).
		WithArgs(evalID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tx, err := db.Begin()
	require.NoError(t, err)

	err = r.SubmitEval(ctx, tx, evalID, profileID,
		[]repo.CompetencyUpsert{{CompetencyID: compID, Rating: 4, Comments: "good"}},
		nil, "en_progreso", setSelf, setRH)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubmitEval_SelfUpsertTargetsSourceConflict(t *testing.T) {
	t.Parallel()
	runSubmitEval(t, "self", true, false)
}

func TestSubmitEval_RHUpsertTargetsSourceConflict(t *testing.T) {
	t.Parallel()
	runSubmitEval(t, "rh", false, true)
}

// TestSubmitEval_ReUpsertUpdatesChangedValue covers reincidence: the second
// upsert for the same (evaluation, competency, source) hits the conflict
// target and updates rating/comments instead of inserting a duplicate.
func TestSubmitEval_ReUpsertUpdatesChangedValue(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewEvaluationRepo(nil, db)
	ctx := context.Background()

	evalID := uuid.New()
	compID := uuid.New()
	profileID := uuid.New()

	submit := func(rating int, comments string) {
		mock.ExpectBegin()
		stubLockRow(mock, evalID)
		mock.ExpectExec(`INSERT INTO evaluation_competencies.+ON CONFLICT \(evaluation_id, competency_id, source\)`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				evalID, compID, rating, comments, profileID, "self", rating).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`UPDATE evaluations SET`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), evalID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`UPDATE evaluations SET version = version \+ 1`).
			WithArgs(evalID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		tx, err := db.Begin()
		require.NoError(t, err)
		err = r.SubmitEval(ctx, tx, evalID, profileID,
			[]repo.CompetencyUpsert{{CompetencyID: compID, Rating: rating, Comments: comments}},
			nil, "en_progreso", true, false)
		require.NoError(t, err)
	}

	submit(3, "first")
	submit(5, "updated")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubmitEval_NilProfileRejected(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewEvaluationRepo(nil, db)
	ctx := context.Background()

	evalID := uuid.New()

	mock.ExpectBegin()
	stubLockRow(mock, evalID)

	tx, err := db.Begin()
	require.NoError(t, err)

	err = r.SubmitEval(ctx, tx, evalID, uuid.Nil,
		[]repo.CompetencyUpsert{{CompetencyID: uuid.New(), Rating: 4}},
		nil, "en_progreso", true, false)
	require.ErrorIs(t, err, repo.ErrEvaluationProfileMissing)
	_ = tx.Rollback()
}

func TestCompetencyRatingRepo_BulkUpsertTargetsSourceConflict(t *testing.T) {
	t.Parallel()

	for _, source := range []string{"self", "rh"} {
		source := source
		t.Run(source, func(t *testing.T) {
			t.Parallel()

			db, mock := newMockDB(t)
			r := repo.NewCompetencyRatingRepo(nil)
			ctx := context.Background()

			evalID := uuid.New()
			compID := uuid.New()
			profileID := uuid.New()

			mock.ExpectBegin()
			mock.ExpectExec(`INSERT INTO evaluation_competencies.+ON CONFLICT \(evaluation_id, competency_id, source\)`).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					evalID, compID, 5, "", profileID, source).
				WillReturnResult(sqlmock.NewResult(0, 1))

			tx, err := db.Begin()
			require.NoError(t, err)

			err = r.BulkUpsert(ctx, tx, evalID, profileID, source,
				[]repo.CompetencyUpsert{{CompetencyID: compID, Rating: 5}})
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestCompetencyRatingRepo_BulkUpsertReUpsertUpdatesChangedValue covers
// reincidence for BulkUpsert: same (evaluation, competency, source) with a
// changed value resolves via the conflict target.
func TestCompetencyRatingRepo_BulkUpsertReUpsertUpdatesChangedValue(t *testing.T) {
	t.Parallel()

	db, mock := newMockDB(t)
	r := repo.NewCompetencyRatingRepo(nil)
	ctx := context.Background()

	evalID := uuid.New()
	compID := uuid.New()
	profileID := uuid.New()

	for _, rating := range []int{2, 5} {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO evaluation_competencies.+ON CONFLICT \(evaluation_id, competency_id, source\)`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				evalID, compID, rating, "", profileID, "rh").
			WillReturnResult(sqlmock.NewResult(0, 1))

		tx, err := db.Begin()
		require.NoError(t, err)

		err = r.BulkUpsert(ctx, tx, evalID, profileID, "rh",
			[]repo.CompetencyUpsert{{CompetencyID: compID, Rating: rating}})
		require.NoError(t, err)
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompetencyRatingRepo_BulkUpsertNilProfileRejected(t *testing.T) {
	t.Parallel()

	r := repo.NewCompetencyRatingRepo(nil)

	// Early guard returns before touching Tx: no Tx expected, nil is safe.
	err := r.BulkUpsert(context.Background(), nil, uuid.New(), uuid.Nil, "rh",
		[]repo.CompetencyUpsert{{CompetencyID: uuid.New(), Rating: 5}})
	require.ErrorIs(t, err, repo.ErrEvaluationProfileMissing)
}

func TestCompetencyRatingRepo_BulkUpsertInvalidSourceRejected(t *testing.T) {
	t.Parallel()

	r := repo.NewCompetencyRatingRepo(nil)

	// Early guard returns before touching Tx: no Tx expected, nil is safe.
	err := r.BulkUpsert(context.Background(), nil, uuid.New(), uuid.New(), "boss",
		[]repo.CompetencyUpsert{{CompetencyID: uuid.New(), Rating: 5}})
	require.ErrorIs(t, err, repo.ErrInvalidCompetencySource)
}
