package scoring

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestProgressPercent(t *testing.T) {
	tests := []struct {
		name          string
		currentValue  float64
		targetValue   float64
		baselineValue float64
		direction     string
		want          float64
	}{
		// ascendente cases
		{name: "ascendente partial", currentValue: 50, targetValue: 100, baselineValue: 0, direction: "ascendente", want: 50},
		{name: "ascendente complete", currentValue: 100, targetValue: 100, baselineValue: 0, direction: "ascendente", want: 100},
		{name: "ascendente over target", currentValue: 150, targetValue: 100, baselineValue: 0, direction: "ascendente", want: 100},
		{name: "ascendente zero current", currentValue: 0, targetValue: 100, baselineValue: 0, direction: "ascendente", want: 0},
		{name: "ascendente zero target", currentValue: 50, targetValue: 0, baselineValue: 0, direction: "ascendente", want: 0},

		// descendente cases
		{name: "descendente partial", currentValue: 30, targetValue: 10, baselineValue: 100, direction: "descendente", want: 70.0 / 90.0 * 100},
		{name: "descendente complete", currentValue: 10, targetValue: 10, baselineValue: 100, direction: "descendente", want: 100},
		{name: "descendente no progress", currentValue: 100, targetValue: 10, baselineValue: 100, direction: "descendente", want: 0},
		{name: "descendente worse than baseline", currentValue: 120, targetValue: 10, baselineValue: 100, direction: "descendente", want: 0},
		{name: "descendente baseline equals target", currentValue: 50, targetValue: 100, baselineValue: 100, direction: "descendente", want: 0},
		{name: "descendente over-target clamp", currentValue: 5, targetValue: 10, baselineValue: 100, direction: "descendente", want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProgressPercent(tt.currentValue, tt.targetValue, tt.baselineValue, tt.direction)
			if !approxEqual(got, tt.want) {
				t.Errorf("ProgressPercent(%v, %v, %v, %q) = %v, want %v",
					tt.currentValue, tt.targetValue, tt.baselineValue, tt.direction, got, tt.want)
			}
		})
	}
}
