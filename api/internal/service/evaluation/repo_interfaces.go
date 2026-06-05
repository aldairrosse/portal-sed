package evaluation

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
)

// --- Repository interfaces ---

// EvaluationRepo defines the operations required by EvaluationService and DashboardService.
type EvaluationRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*repo.EvaluationRow, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	LockEvalForUpdate(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) (*repo.EvaluationRow, error)
	SubmitEval(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, comps []repo.CompetencyUpsert, goals []repo.GoalCommentUpsert, newState string, setSelfCompleted, setRHCompleted bool) error
	GetDetail(ctx context.Context, id uuid.UUID) (*repo.EvaluationRow, []*internal.EvaluationCompetency, []*internal.EvaluationGoal, error)
	ListByCycle(ctx context.Context, cycleID uuid.UUID, state string, cursor string, limit int) ([]*repo.EvaluationRow, string, error)
	FinalizeEval(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) error
	RefreshSummaryView(ctx context.Context) error
	GetSummaryByCycle(ctx context.Context, cycleID uuid.UUID) (map[string]int64, error)
}

// CompetencyRatingRepo defines the operations required for competency ratings.
type CompetencyRatingRepo interface {
	BulkUpsert(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, comps []repo.CompetencyUpsert) error
	DeleteByEvaluation(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) error
	GetByEvaluation(ctx context.Context, evalID uuid.UUID) ([]*internal.EvaluationCompetency, error)
}

// GoalRatingRepo defines the operations required for goal comments.
type GoalRatingRepo interface {
	UpdateComments(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, goals []repo.GoalCommentUpsert) error
	GetByEvaluation(ctx context.Context, evalID uuid.UUID) ([]*internal.EvaluationGoal, error)
	VerifyGoalsExist(ctx context.Context, evalID uuid.UUID, goalIDs []uuid.UUID) error
}

// NineBoxRepo defines the operations required for 9×9 matrices and entries.
type NineBoxRepo interface {
	CreateMatrix(ctx context.Context, cycleID, evaluatorID uuid.UUID) (*internal.NineBoxMatrix, error)
	GetMatrixByID(ctx context.Context, id uuid.UUID) (*internal.NineBoxMatrix, error)
	ListMatrices(ctx context.Context, cycleID, evaluatorID uuid.UUID) ([]*internal.NineBoxMatrix, error)
	GetMatrixEntries(ctx context.Context, matrixID uuid.UUID) ([]*internal.NineBoxEntry, error)
	UpsertEntry(ctx context.Context, tx *sql.Tx, matrixID uuid.UUID, evaluateeID uuid.UUID, perf, pot int, quadrant int, comments string) (*internal.NineBoxEntry, error)
	UpdateEntry(ctx context.Context, tx *sql.Tx, entryID uuid.UUID, perf, pot int, quadrant int, comments string, version int) (*internal.NineBoxEntry, error)
	BatchUpsertEntries(ctx context.Context, tx *sql.Tx, matrixID uuid.UUID, items []repo.EntryUpsert) ([]*internal.NineBoxEntry, error)
	LockEntryForSelect(ctx context.Context, tx *sql.Tx, matrixID, evaluateeID uuid.UUID) error
	FetchEntryVersion(ctx context.Context, entryID uuid.UUID) (int, error)
}

// CatalogRepo defines read-only catalog operations for scales and quadrants.
type CatalogRepo interface {
	GetQuadrants(ctx context.Context) ([]*internal.NineBoxQuadrant, error)
	GetQuadrantByNumber(ctx context.Context, quadrant int) (*internal.NineBoxQuadrant, error)
	GetScales(ctx context.Context) ([]*internal.NineBoxScale, error)
}

// DB is the minimal database interface needed to begin transactions.
type DB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
