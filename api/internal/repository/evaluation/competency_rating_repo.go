package evaluation

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluationcompetency"
)

// CompetencyRatingRepo handles bulk operations for EvaluationCompetency ratings.
type CompetencyRatingRepo struct {
	client *internal.Client
}

// NewCompetencyRatingRepo creates a new CompetencyRatingRepo.
func NewCompetencyRatingRepo(client *internal.Client) *CompetencyRatingRepo {
	return &CompetencyRatingRepo{client: client}
}

// BulkUpsert performs an atomic upsert of multiple competency ratings within a transaction.
func (r *CompetencyRatingRepo) BulkUpsert(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, comps []CompetencyUpsert) error {
	if len(comps) == 0 {
		return nil
	}

	now := time.Now()
	for _, c := range comps {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO evaluation_competencies (id, created_at, updated_at, evaluation_id, competency_id, rating, comments, profile_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (evaluation_id, competency_id) DO UPDATE
			 SET rating = EXCLUDED.rating, comments = EXCLUDED.comments, updated_at = EXCLUDED.updated_at`,
			uuid.New(), now, now, evalID, c.CompetencyID, c.Rating, c.Comments, uuid.Nil,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteByEvaluation removes all competency ratings for a given evaluation within a tx.
func (r *CompetencyRatingRepo) DeleteByEvaluation(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM evaluation_competencies WHERE evaluation_id = $1`, evalID,
	)
	return err
}

// GetByEvaluation retrieves all competency ratings for an evaluation.
func (r *CompetencyRatingRepo) GetByEvaluation(ctx context.Context, evalID uuid.UUID) ([]*internal.EvaluationCompetency, error) {
	results, err := r.client.EvaluationCompetency.Query().
		Where(evaluationcompetency.EvaluationID(evalID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []*internal.EvaluationCompetency{}, nil
	}
	return results, nil
}
