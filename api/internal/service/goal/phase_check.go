package goal

import (
	"context"
	"fmt"

	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/state"
)

// CyclePhase represents a phase in the evaluation cycle.
type CyclePhase string

const (
	PhaseAsignacion CyclePhase = "asignacion"
	PhaseAvance     CyclePhase = "avance"
	// PhaseMedioAnio is the mid-year phase; kept compatible with PhaseAvance:
	// both names refer to the same mid-year phase for write gates.
	PhaseMedioAnio CyclePhase = "medio-anio"
	PhaseCierre    CyclePhase = "cierre"
)

// PhaseChecker provides the current phase for an employee's active cycle.
// This is implemented by the C2 (evaluation-lifecycle-api) integration.
type PhaseChecker interface {
	// GetCurrentPhase returns the current phase for the employee's active cycle.
	GetCurrentPhase(ctx context.Context, empID string) (CyclePhase, error)
}

// PhaseCheck provides phase enforcement for goals operations.
type PhaseCheck struct {
	checker PhaseChecker
}

// NewPhaseCheck creates a new PhaseCheck.
func NewPhaseCheck(checker PhaseChecker) *PhaseCheck {
	return &PhaseCheck{checker: checker}
}

// ErrPhaseRestricted is returned when an operation is not allowed in the current phase.
var ErrPhaseRestricted = pkgerrors.ErrPhaseRestricted

// CurrentPhase returns the current phase for the employee's active cycle.
func (pc *PhaseCheck) CurrentPhase(ctx context.Context, empID string) (CyclePhase, error) {
	return pc.checker.GetCurrentPhase(ctx, empID)
}

// isMidPhase reports whether p is the mid-year phase ("avance"/"medio-anio").
// Delegates to state.IsMidYearPhase (single source of truth).
func isMidPhase(p CyclePhase) bool {
	return state.IsMidYearPhase(string(p))
}

// samePhaseForWrite treats "avance" and "medio-anio" as the same phase.
// Delegates to state.SamePhaseForWrite (single source of truth).
func samePhaseForWrite(a, b CyclePhase) bool {
	return state.SamePhaseForWrite(string(a), string(b))
}

// Enforce checks that the current phase is one of the allowed phases.
// "avance" and "medio-anio" are interchangeable.
func (pc *PhaseCheck) Enforce(ctx context.Context, empID string, allowed ...CyclePhase) error {
	phase, err := pc.checker.GetCurrentPhase(ctx, empID)
	if err != nil {
		return fmt.Errorf("phase check: %w", err)
	}
	for _, p := range allowed {
		if samePhaseForWrite(p, phase) {
			return nil
		}
	}
	return ErrPhaseRestricted
}

// CanCreateGoal checks if the current phase allows goal creation.
func (pc *PhaseCheck) CanCreateGoal(ctx context.Context, empID string) error {
	return pc.Enforce(ctx, empID, PhaseAsignacion)
}

// CanUpdateGoal checks if the current phase allows goal updates.
func (pc *PhaseCheck) CanUpdateGoal(ctx context.Context, empID string) error {
	return pc.Enforce(ctx, empID, PhaseAsignacion)
}

// CanDeleteGoal checks if the current phase allows goal deletion.
func (pc *PhaseCheck) CanDeleteGoal(ctx context.Context, empID string) error {
	return pc.Enforce(ctx, empID, PhaseAsignacion)
}

// CanUpdateProgress checks if the current phase allows progress updates.
func (pc *PhaseCheck) CanUpdateProgress(ctx context.Context, empID string) error {
	return pc.Enforce(ctx, empID, PhaseAvance, PhaseMedioAnio)
}

// CanCreateCategory checks if the current phase allows category creation.
func (pc *PhaseCheck) CanCreateCategory(ctx context.Context, empID string) error {
	return pc.Enforce(ctx, empID, PhaseAsignacion)
}

// CanUpdateCategory checks if the current phase allows category updates.
func (pc *PhaseCheck) CanUpdateCategory(ctx context.Context, empID string) error {
	return pc.Enforce(ctx, empID, PhaseAsignacion, PhaseAvance, PhaseMedioAnio)
}

// CanUpdateCategoryField checks if a specific field can be updated in the current phase.
// In avance phase, only weight changes are allowed for categories.
func (pc *PhaseCheck) CanUpdateCategoryField(ctx context.Context, empID string, fieldName string) error {
	phase, err := pc.checker.GetCurrentPhase(ctx, empID)
	if err != nil {
		return fmt.Errorf("phase check: %w", err)
	}
	if isMidPhase(phase) && fieldName != "weight" {
		return ErrPhaseRestricted
	}
	return pc.Enforce(ctx, empID, PhaseAsignacion, PhaseAvance, PhaseMedioAnio)
}

// CanDeleteCategory checks if the current phase allows category deletion.
func (pc *PhaseCheck) CanDeleteCategory(ctx context.Context, empID string) error {
	return pc.Enforce(ctx, empID, PhaseAsignacion)
}

// CanLinkKPI checks if the current phase allows KPI linking.
func (pc *PhaseCheck) CanLinkKPI(ctx context.Context, empID string) error {
	return pc.Enforce(ctx, empID, PhaseAsignacion)
}

// CanCreateAssignment checks if the current phase allows assignment creation.
func (pc *PhaseCheck) CanCreateAssignment(ctx context.Context, empID string) error {
	return pc.Enforce(ctx, empID, PhaseAsignacion)
}
