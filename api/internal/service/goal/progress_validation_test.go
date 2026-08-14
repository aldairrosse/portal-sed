package goal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func floatPtr(v float64) *float64 {
	return &v
}

func TestValidateProgressValue(t *testing.T) {
	tests := []struct {
		name          string
		unit          string
		direction     string
		baselineValue *float64
		currentValue  float64
		wantErr       bool
	}{
		{name: "binario 0 ok", unit: "binario", direction: "ascendente", baselineValue: nil, currentValue: 0},
		{name: "binario 1 ok", unit: "binario", direction: "ascendente", baselineValue: nil, currentValue: 1},
		{name: "binario 0.5 error", unit: "binario", direction: "ascendente", baselineValue: nil, currentValue: 0.5, wantErr: true},
		{name: "descendente igual baseline ok", unit: "numero", direction: "descendente", baselineValue: floatPtr(100), currentValue: 100},
		{name: "descendente menor baseline ok", unit: "numero", direction: "descendente", baselineValue: floatPtr(100), currentValue: 50},
		{name: "descendente mayor baseline error", unit: "numero", direction: "descendente", baselineValue: floatPtr(100), currentValue: 150, wantErr: true},
		{name: "negativo error", unit: "numero", direction: "ascendente", baselineValue: nil, currentValue: -1, wantErr: true},
		{name: "ascendente sin restriccion ok", unit: "numero", direction: "ascendente", baselineValue: nil, currentValue: 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProgressValue(tt.unit, tt.direction, tt.baselineValue, tt.currentValue)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
