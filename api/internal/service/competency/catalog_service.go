package competency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/competency"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/competency"
)

// catalogService implements CatalogService.
type catalogService struct {
	repo repo.CatalogRepo
}

// NewCatalogService creates a new CatalogService.
func NewCatalogService(repo repo.CatalogRepo) CatalogService {
	return &catalogService{repo: repo}
}

func (s *catalogService) ListLevels(ctx context.Context) ([]dto.LevelDefinitionItem, error) {
	levels, err := s.repo.ListLevels(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]dto.LevelDefinitionItem, len(levels))
	for i, l := range levels {
		items[i] = dto.LevelDefinitionItem{
			Level:       l.Level,
			Label:       l.Label,
			Description: l.Description,
		}
	}
	return items, nil
}

func (s *catalogService) ListProfiles(ctx context.Context) ([]dto.EvaluationProfileItem, error) {
	profiles, err := s.repo.ListProfiles(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]dto.EvaluationProfileItem, len(profiles))
	for i, p := range profiles {
		items[i] = dto.EvaluationProfileItem{
			ID:          p.ID.String(),
			Name:        p.Name,
			Description: p.Description,
		}
	}
	return items, nil
}

// ComputeETag computes an ETag string from a serialisable value by
// JSON-marshalling it and taking the SHA256 hex digest.
func ComputeETag(prefix string, v interface{}) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(raw)
	return fmt.Sprintf("%s:%s", prefix, hex.EncodeToString(h[:])), nil
}
