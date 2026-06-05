package cycle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
)

// PhaseDefinitionResponse is the API response for a phase definition.
type PhaseDefinitionResponse struct {
	Phase          string   `json:"phase"`
	Label          string   `json:"label"`
	Order          int      `json:"order"`
	AllowedActors  []string `json:"allowed_actors"`
	AllowedActions []string `json:"allowed_actions"`
	BlockedActions []string `json:"blocked_actions"`
}

// PhaseTransitionResponse is the API response for a phase transition.
type PhaseTransitionResponse struct {
	FromPhase  string                 `json:"from_phase"`
	ToPhase    string                 `json:"to_phase"`
	Trigger    string                 `json:"trigger"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
}

// PhaseService defines the interface for phase-related read operations.
type PhaseService interface {
	GetPhaseDefinitions(ctx context.Context) ([]*PhaseDefinitionResponse, string, error)
	GetAvailableTransitions(ctx context.Context, cycleID string) ([]*PhaseTransitionResponse, error)
}

// phaseService implements PhaseService.
type phaseService struct {
	cycleRepo *repo.CycleRepo
	phaseRepo *repo.PhaseRepo
}

// NewPhaseService creates a new phase service.
func NewPhaseService(cycleRepo *repo.CycleRepo, phaseRepo *repo.PhaseRepo) PhaseService {
	return &phaseService{
		cycleRepo: cycleRepo,
		phaseRepo: phaseRepo,
	}
}

// GetPhaseDefinitions returns all phase definitions with an ETag computed as
// SHA256 of the JSON payload (first 16 hex chars).
func (s *phaseService) GetPhaseDefinitions(ctx context.Context) ([]*PhaseDefinitionResponse, string, error) {
	rows, err := s.phaseRepo.GetPhaseDefinitions(ctx)
	if err != nil {
		return nil, "", err
	}

	resp := make([]*PhaseDefinitionResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, &PhaseDefinitionResponse{
			Phase:          r.Phase,
			Label:          r.Label,
			Order:          r.Order,
			AllowedActors:  r.AllowedActors,
			AllowedActions: r.AllowedActions,
			BlockedActions: r.BlockedActions,
		})
	}

	// Compute ETag
	etag, err := computeETag(resp)
	if err != nil {
		return nil, "", err
	}

	return resp, etag, nil
}

// GetAvailableTransitions returns the available transitions from the current
// phase of the given cycle.
func (s *phaseService) GetAvailableTransitions(ctx context.Context, cycleID string) ([]*PhaseTransitionResponse, error) {
	id, err := uuid.Parse(cycleID)
	if err != nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cycle_id must be a valid UUID v4", err)
	}

	// Fetch cycle to get current_phase
	row, err := s.cycleRepo.GetCycle(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get transitions from current phase
	transitions, err := s.phaseRepo.GetTransitionsByFromPhase(ctx, string(row.CurrentPhase))
	if err != nil {
		return nil, err
	}

	resp := make([]*PhaseTransitionResponse, 0, len(transitions))
	for _, t := range transitions {
		resp = append(resp, &PhaseTransitionResponse{
			FromPhase:  t.FromPhase,
			ToPhase:    t.ToPhase,
			Trigger:    t.Trigger,
			Conditions: t.Conditions,
		})
	}

	return resp, nil
}

// computeETag computes an ETag as the first 16 hex chars of SHA256(jsonPayload).
func computeETag(v interface{}) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:8]), nil // first 8 bytes = 16 hex chars
}
