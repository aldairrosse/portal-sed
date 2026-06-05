package competency

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/scalecriterion"
)

// scaleRepo implements ScaleRepo.
type scaleRepo struct {
	client *internal.Client
}

// NewScaleRepo creates a new ScaleRepo.
func NewScaleRepo(client *internal.Client) ScaleRepo {
	return &scaleRepo{client: client}
}

func (r *scaleRepo) WithTx(ctx context.Context, fn TxFunc) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func (r *scaleRepo) GetByCompetency(ctx context.Context, competencyID string) ([]*internal.ScaleCriterion, error) {
	criteria, err := r.client.ScaleCriterion.Query().
		Where(scalecriterion.CompetencyIDEQ(uuid.MustParse(competencyID))).
		Order(internal.Asc(scalecriterion.FieldLevel)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return criteria, nil
}

// ReplaceAll runs inside a transaction: deletes all existing criteria for the
// competency, bulk-inserts the new ones, and bumps the competency's updated_at.
func (r *scaleRepo) ReplaceAll(ctx context.Context, competencyID, pillarID string, criteria []ScaleCriterionInput) error {
	return r.WithTx(ctx, func(tx *internal.Tx) error {
		compID := uuid.MustParse(competencyID)
		pID := uuid.MustParse(pillarID)

		// Delete existing criteria
		if _, err := tx.ScaleCriterion.Delete().
			Where(scalecriterion.CompetencyIDEQ(compID)).
			Exec(ctx); err != nil {
			return err
		}

		// Bulk insert new criteria
		if len(criteria) > 0 {
			builders := make([]*internal.ScaleCriterionCreate, len(criteria))
			for i, c := range criteria {
				builders[i] = tx.ScaleCriterion.Create().
					SetCompetencyID(compID).
					SetPillarID(pID).
					SetLevel(c.Level).
					SetDescription(c.Description)
			}
			if _, err := tx.ScaleCriterion.CreateBulk(builders...).Save(ctx); err != nil {
				return err
			}
		}

		// Bump competency updated_at
		if err := tx.Competency.UpdateOneID(compID).
			SetUpdatedAt(time.Now()).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}
