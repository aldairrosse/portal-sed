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

// RequiresPhase validates that the current phase is "cierre". Returns an
// INVALID_PHASE domain error if the phase does not match.
func RequiresPhase(phase string) error {
	if phase != "cierre" {
		return pkgerrors.NewDomainError(
			pkgerrors.PhaseNotAdvanceable,
			fmt.Sprintf("this operation requires the cycle to be in 'cierre' phase; current phase is '%s'", phase),
			nil,
		)
	}
	return nil
}

// String returns the string representation of the state.
func (s EvaluationState) String() string {
	return string(s)
}
