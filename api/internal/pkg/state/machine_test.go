package state_test

import (
	"testing"

	"github.com/sed-evaluacion-desempeno/api/internal/pkg/state"
	"github.com/stretchr/testify/assert"
)

func TestCanTransition_AllValid(t *testing.T) {
	valid := []struct {
		from state.EvaluationState
		to   state.EvaluationState
	}{
		{state.StatePendingEvalFinal, state.StateInProgress},
		{state.StateInProgress, state.StateCompleted},
	}

	for _, tt := range valid {
		t.Run("", func(t *testing.T) {
			assert.True(t, state.CanTransition(tt.from, tt.to),
				"expected %s → %s to be valid", tt.from, tt.to)
		})
	}
}

func TestCanTransition_AllInvalid(t *testing.T) {
	invalid := []struct {
		from state.EvaluationState
		to   state.EvaluationState
	}{
		{state.StatePendingEvalFinal, state.StateCompleted},
		{state.StatePendingEvalFinal, state.StatePendingEvalFinal},
		{state.StateInProgress, state.StateInProgress},
		{state.StateInProgress, state.StatePendingEvalFinal},
		{state.StateCompleted, state.StatePendingEvalFinal},
		{state.StateCompleted, state.StateInProgress},
		{state.StateCompleted, state.StateCompleted},
	}

	for _, tt := range invalid {
		t.Run("", func(t *testing.T) {
			assert.False(t, state.CanTransition(tt.from, tt.to),
				"expected %s → %s to be invalid", tt.from, tt.to)
		})
	}
}

func TestCanTransition_PendingToInProgress(t *testing.T) {
	assert.True(t, state.CanTransition(state.StatePendingEvalFinal, state.StateInProgress))
}

func TestCanTransition_CompletedIsTerminal(t *testing.T) {
	assert.True(t, state.IsTerminal(state.StateCompleted))
	assert.False(t, state.CanTransition(state.StateCompleted, state.StateInProgress))
	assert.False(t, state.CanTransition(state.StateCompleted, state.StatePendingEvalFinal))
}

func TestCanTransition_CannotGoBack(t *testing.T) {
	assert.False(t, state.CanTransition(state.StateInProgress, state.StatePendingEvalFinal),
		"cannot revert from in_progress to pending")
	assert.False(t, state.CanTransition(state.StateCompleted, state.StateInProgress),
		"cannot revert from completed to in_progress")
	assert.False(t, state.CanTransition(state.StateCompleted, state.StatePendingEvalFinal),
		"cannot revert from completed to pending")
}

func TestIsMidYearPhase(t *testing.T) {
	tests := []struct {
		name  string
		phase string
		want  bool
	}{
		{"avance is mid-year", state.PhaseAvance, true},
		{"medio-anio is mid-year", state.PhaseMedioAnio, true},
		{"asignacion is not mid-year", state.PhaseAsignacion, false},
		{"cierre is not mid-year", state.PhaseCierre, false},
		{"unknown is not mid-year", "desconocida", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, state.IsMidYearPhase(tt.phase))
		})
	}
}

func TestNormalizePhase(t *testing.T) {
	tests := []struct {
		name  string
		phase string
		want  string
	}{
		{"avance canonical", "avance", state.PhaseAvance},
		{"cierre canonical", "cierre", state.PhaseCierre},
		{"asignacion canonical", "asignacion", state.PhaseAsignacion},
		{"medio-anio alias", "medio-anio", state.PhaseAvance},
		{"medio_anio alias", "medio_anio", state.PhaseAvance},
		{"medioanio alias", "medioanio", state.PhaseAvance},
		{"fin-anio alias", "fin-anio", state.PhaseCierre},
		{"fin_anio alias", "fin_anio", state.PhaseCierre},
		{"finanio alias", "finanio", state.PhaseCierre},
		{"inicio-anio alias", "inicio-anio", state.PhaseAsignacion},
		{"inicio_anio alias", "inicio_anio", state.PhaseAsignacion},
		{"case-insensitive", "MEDIO-ANIO", state.PhaseAvance},
		{"case-insensitive avance", "AVANCE", state.PhaseAvance},
		{"trim spaces", "  avance  ", state.PhaseAvance},
		{"trim + case alias", "  Medio_Anio ", state.PhaseAvance},
		{"unknown passthrough", "desconocida", "desconocida"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, state.NormalizePhase(tt.phase))
		})
	}
}

func TestSamePhaseForWrite(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"same phase", state.PhaseAsignacion, state.PhaseAsignacion, true},
		{"avance to medio-anio read-only equivalent", state.PhaseAvance, state.PhaseMedioAnio, true},
		{"medio-anio to avance read-only equivalent", state.PhaseMedioAnio, state.PhaseAvance, true},
		{"asignacion vs avance differ", state.PhaseAsignacion, state.PhaseAvance, false},
		{"avance vs cierre differ", state.PhaseAvance, state.PhaseCierre, false},
		{"cierre vs asignacion differ", state.PhaseCierre, state.PhaseAsignacion, false},
		{"medio_anio to avance alias", "medio_anio", state.PhaseAvance, true},
		{"medioanio to avance alias", "medioanio", state.PhaseAvance, true},
		{"fin-anio to cierre alias", "fin-anio", state.PhaseCierre, true},
		{"finanio to cierre alias", "finanio", state.PhaseCierre, true},
		{"inicio-anio to asignacion alias", "inicio-anio", state.PhaseAsignacion, true},
		{"case-insensitive match", "AVANCE", "medio-anio", true},
		{"trimmed match", "  avance  ", "medio_anio", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, state.SamePhaseForWrite(tt.a, tt.b))
		})
	}
}

func TestBatchTransition_MultipleEvaluations(t *testing.T) {
	evals := []state.EvaluationState{
		state.StatePendingEvalFinal,
		state.StatePendingEvalFinal,
		state.StateInProgress,
	}

	// Simulate advancing all pending evaluations to in_progress
	for i, s := range evals {
		if state.CanTransition(s, state.StateInProgress) {
			evals[i] = state.StateInProgress
		}
	}

	assert.Equal(t, state.StateInProgress, evals[0])
	assert.Equal(t, state.StateInProgress, evals[1])
	assert.Equal(t, state.StateInProgress, evals[2],
		"in_progress should remain in_progress when transition is attempted")
}
