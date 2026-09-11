package goal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubChecker struct{ phase CyclePhase }

func (s stubChecker) GetCurrentPhase(_ context.Context, _ string) (CyclePhase, error) {
	return s.phase, nil
}

func TestCanUpdateProgress(t *testing.T) {
	tests := []struct {
		name    string
		phase   CyclePhase
		wantErr bool
	}{
		{name: "avance ok", phase: PhaseAvance},
		{name: "medio-anio alias ok", phase: PhaseMedioAnio},
		{name: "cierre ok", phase: PhaseCierre},
		{name: "asignacion bloqueado", phase: PhaseAsignacion, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := NewPhaseCheck(stubChecker{phase: tt.phase})
			err := pc.CanUpdateProgress(context.Background(), "emp-1")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
