package competency

import (
	"context"
	"database/sql"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluationprofile"
	"github.com/sed-evaluacion-desempeno/api/internal/leveldefinition"
)

// catalogRepo implements CatalogRepo.
type catalogRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewCatalogRepo creates a new CatalogRepo.
func NewCatalogRepo(client *internal.Client, db *sql.DB) CatalogRepo {
	return &catalogRepo{client: client, db: db}
}

func (r *catalogRepo) ListLevels(ctx context.Context) ([]*internal.LevelDefinition, error) {
	return r.client.LevelDefinition.Query().
		Order(internal.Asc(leveldefinition.FieldLevel)).
		All(ctx)
}

// UpdateLevel upserts a level definition. Ent does not expose OnConflict(),
// so we use raw SQL with INSERT ... ON CONFLICT (level) DO UPDATE.
func (r *catalogRepo) UpdateLevel(ctx context.Context, level int, label, description string) error {
	const q = `
		INSERT INTO level_definitions (level, label, description)
		VALUES ($1, $2, NULLIF($3, ''))
		ON CONFLICT (level) DO UPDATE
		   SET label = EXCLUDED.label,
		       description = EXCLUDED.description`
	_, err := r.db.ExecContext(ctx, q, level, label, description)
	return err
}

func (r *catalogRepo) ListProfiles(ctx context.Context) ([]*internal.EvaluationProfile, error) {
	return r.client.EvaluationProfile.Query().
		Order(internal.Asc(evaluationprofile.FieldName)).
		All(ctx)
}
