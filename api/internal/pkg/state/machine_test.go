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
