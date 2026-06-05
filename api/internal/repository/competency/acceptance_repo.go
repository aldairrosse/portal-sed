package competency

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/competencyacceptancelevel"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// acceptanceRepo implements AcceptanceRepo.
type acceptanceRepo struct {
	client *internal.Client
}

// NewAcceptanceRepo creates a new AcceptanceRepo.
func NewAcceptanceRepo(client *internal.Client) AcceptanceRepo {
	return &acceptanceRepo{client: client}
}

func (r *acceptanceRepo) WithTx(ctx context.Context, fn TxFunc) error {
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
			err = pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"transaction rollback failed", rerr)
		}
		return err
	}
	return tx.Commit()
}

func (r *acceptanceRepo) List(ctx context.Context, competencyID, profileID *string) ([]*internal.CompetencyAcceptanceLevel, error) {
	q := r.client.CompetencyAcceptanceLevel.Query()

	if competencyID != nil && *competencyID != "" {
		q = q.Where(competencyacceptancelevel.CompetencyIDEQ(uuid.MustParse(*competencyID)))
	}
	if profileID != nil && *profileID != "" {
		q = q.Where(competencyacceptancelevel.ProfileIDEQ(uuid.MustParse(*profileID)))
	}

	return q.All(ctx)
}

func (r *acceptanceRepo) Upsert(ctx context.Context, competencyID, profileID string, level int) (*internal.CompetencyAcceptanceLevel, error) {
	compID := uuid.MustParse(competencyID)
	profID := uuid.MustParse(profileID)

	// Check if existing row exists
	exists, err := r.client.CompetencyAcceptanceLevel.Query().
		Where(
			competencyacceptancelevel.CompetencyIDEQ(compID),
			competencyacceptancelevel.ProfileIDEQ(profID),
		).
		Exist(ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		// Create
		created, err := r.client.CompetencyAcceptanceLevel.Create().
			SetCompetencyID(compID).
			SetProfileID(profID).
			SetLevel(level).
			Save(ctx)
		if err != nil {
			if internal.IsConstraintError(err) {
				return nil, pkgerrors.NewDomainError("DUPLICATE_ENTRY",
					"duplicate acceptance level entry", err)
			}
			return nil, err
		}
		return created, nil
	}

	// Update existing
	if _, err := r.client.CompetencyAcceptanceLevel.Update().
		Where(
			competencyacceptancelevel.CompetencyIDEQ(compID),
			competencyacceptancelevel.ProfileIDEQ(profID),
		).
		SetLevel(level).
		SetUpdatedAt(time.Now()).
		Save(ctx); err != nil {
		return nil, err
	}

	// Re-fetch to return full entity
	return r.client.CompetencyAcceptanceLevel.Query().
		Where(
			competencyacceptancelevel.CompetencyIDEQ(compID),
			competencyacceptancelevel.ProfileIDEQ(profID),
		).
		Only(ctx)
}
