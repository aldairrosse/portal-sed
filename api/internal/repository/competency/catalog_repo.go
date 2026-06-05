package competency

import (
	"context"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluationprofile"
	"github.com/sed-evaluacion-desempeno/api/internal/leveldefinition"
)

// catalogRepo implements CatalogRepo.
type catalogRepo struct {
	client *internal.Client
}

// NewCatalogRepo creates a new CatalogRepo.
func NewCatalogRepo(client *internal.Client) CatalogRepo {
	return &catalogRepo{client: client}
}

func (r *catalogRepo) ListLevels(ctx context.Context) ([]*internal.LevelDefinition, error) {
	return r.client.LevelDefinition.Query().
		Order(internal.Asc(leveldefinition.FieldLevel)).
		All(ctx)
}

func (r *catalogRepo) ListProfiles(ctx context.Context) ([]*internal.EvaluationProfile, error) {
	return r.client.EvaluationProfile.Query().
		Order(internal.Asc(evaluationprofile.FieldName)).
		All(ctx)
}
