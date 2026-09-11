// Package state provides a lightweight, pure state transition guard for
// Evaluation states during the year-end closing phase.
//
// Valid transitions (during "cierre" phase):
//   - pendiente_evaluacion_final → en_progreso   (first submission by employee or RH)
//   - en_progreso               → completada     (finalization by RH)
//
// Only "completada" is terminal.
package state

import (
	"fmt"
	"strings"

	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// EvaluationState represents the state of an evaluation.
type EvaluationState string

const (
	// StatePendingEvalFinal is the initial state when the cycle reaches "cierre".
	// Self-evaluation and RH evaluation have not been submitted yet.
	StatePendingEvalFinal EvaluationState = "pendiente_evaluacion_final"

	// StateInProgress means at least one evaluation path (self or RH) has been
	// submitted, but the evaluation is not yet finalized.
	StateInProgress EvaluationState = "en_progreso"

	// StateCompleted means the evaluation has been finalized by RH and no
	// further changes are allowed.
	StateCompleted EvaluationState = "completada"
)

// Cycle phase names (mirrors cycles.current_phase enum).
// Canonical: asignacion, avance, cierre. medio-anio is a read-only alias
// of avance for legacy cycles.
const (
	PhaseAsignacion = "asignacion"
	PhaseAvance     = "avance"
	// PhaseMedioAnio is a deprecated read alias of avance (see IsMidYearPhase).
	PhaseMedioAnio = "medio-anio"
	PhaseCierre    = "cierre"
	// PhaseFinAnio is a deprecated read alias of cierre.
	PhaseFinAnio = "fin-anio"
	// PhaseInicioAnio is a deprecated read alias of asignacion.
	PhaseInicioAnio = "inicio-anio"
)

// validTransitions defines the allowed state transitions.
var validTransitions = map[EvaluationState][]EvaluationState{
	StatePendingEvalFinal: {StateInProgress},
	StateInProgress:       {StateCompleted},
}

// IsTerminal returns true if the state is "completada" (no further transitions).
func IsTerminal(state EvaluationState) bool {
	return state == StateCompleted
}

// CanTransition checks whether a transition from `from` to `to` is valid
// according to the state machine.
func CanTransition(from, to EvaluationState) bool {
	if from == StateCompleted {
		return false
	}
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// NormalizePhase lowercases/trims and maps legacy aliases to the canonical
// phase: medio-anio/medio_anio/medioanio → avance,
// fin-anio/fin_anio/finanio → cierre, inicio-anio/inicio_anio → asignacion.
// Canonical avance/cierre/asignacion pass through; anything else returns the
// lower-trimmed input.
func NormalizePhase(phase string) string {
	n := strings.ToLower(strings.TrimSpace(phase))
	switch n {
	case "medio-anio", "medio_anio", "medioanio", PhaseAvance:
		return PhaseAvance
	case "fin-anio", "fin_anio", "finanio", PhaseCierre:
		return PhaseCierre
	case "inicio-anio", "inicio_anio", PhaseAsignacion:
		return PhaseAsignacion
	default:
		return n
	}
}

// normalizePhase kept as thin wrapper (single source: NormalizePhase).
func normalizePhase(p string) string {
	return NormalizePhase(p)
}

// IsMidYearPhase reports whether phase is the mid-year phase.
// "avance" and "medio-anio" are treated as the same mid-year phase.
func IsMidYearPhase(phase string) bool {
	switch NormalizePhase(phase) {
	case PhaseAvance, PhaseMedioAnio:
		return true
	default:
		return false
	}
}

// SamePhaseForWrite reports whether two phase names match for write gates,
// comparing canonical phases (aliases unified via NormalizePhase).
func SamePhaseForWrite(a, b string) bool {
	return NormalizePhase(a) == NormalizePhase(b)
}

// IsWritablePhase reports whether writes are allowed in the given phase
// (mid-year avance/medio-anio or cierre).
func IsWritablePhase(phase string) bool {
	n := NormalizePhase(phase)
	return n == PhaseCierre || IsMidYearPhase(n)
}

// RequiresPhase validates that the current phase is "cierre". Returns a
// PHASE_NOT_ADVANCEABLE domain error if the phase does not match.
// Kept for cierre-only operations (self/RH/finalize); use WritableInPhase
// for mid-year writes.
func RequiresPhase(phase string) error {
	if NormalizePhase(phase) != PhaseCierre {
		return pkgerrors.NewDomainError(
			pkgerrors.PhaseNotAdvanceable,
			fmt.Sprintf("this operation requires the cycle to be in 'cierre' phase; current phase is '%s'", phase),
			nil,
		)
	}
	return nil
}

// WritableInPhase validates that the cycle's current phase matches the
// required phase for a write (mid-year "avance"/"medio-anio" equivalent).
// Returns a PHASE_NOT_ADVANCEABLE (409) domain error on mismatch.
// Permission checks (403) are enforced separately by RBAC middleware.
func WritableInPhase(currentPhase, requiredPhase string) error {
	if requiredPhase == "" {
		if !IsWritablePhase(currentPhase) {
			return pkgerrors.NewDomainError(
				pkgerrors.PhaseNotAdvanceable,
				fmt.Sprintf("writes are only allowed in 'avance'/'medio-anio' or 'cierre' phase; current phase is '%s'", currentPhase),
				nil,
			)
		}
		return nil
	}
	if !SamePhaseForWrite(currentPhase, requiredPhase) {
		return pkgerrors.NewDomainError(
			pkgerrors.PhaseNotAdvanceable,
			fmt.Sprintf("this operation requires the cycle to be in '%s' phase; current phase is '%s'", requiredPhase, currentPhase),
			nil,
		)
	}
	return nil
}

// CyclePhaseOrder is the canonical intra-cycle phase progression.
// medio-anio is NOT a step; it reads as avance (see phaseOrderIndex).
var CyclePhaseOrder = []string{PhaseAsignacion, PhaseAvance, PhaseCierre}

// phaseOrderIndex returns the index of a phase in CyclePhaseOrder, or -1.
// Legacy "medio-anio" maps to the avance index for read compatibility.
func phaseOrderIndex(phase string) int {
	n := NormalizePhase(phase)
	if n == PhaseMedioAnio {
		n = PhaseAvance
	}
	for i, p := range CyclePhaseOrder {
		if p == n {
			return i
		}
	}
	return -1
}

// IsForwardAllowed reports whether toPhase is the immediate next phase after
// fromPhase within an unfinished active cycle.
func IsForwardAllowed(fromPhase, toPhase string, cycleActive bool) bool {
	if !cycleActive {
		return false
	}
	return phaseOrderIndex(toPhase) == phaseOrderIndex(fromPhase)+1
}

// IsBackwardAllowed reports whether toPhase is the immediate previous phase
// before fromPhase within an unfinished active cycle.
func IsBackwardAllowed(fromPhase, toPhase string, cycleActive bool) bool {
	if !cycleActive {
		return false
	}
	return phaseOrderIndex(toPhase) == phaseOrderIndex(fromPhase)-1
}

// String returns the string representation of the state.
func (s EvaluationState) String() string {
	return string(s)
}
