package quadrant_test

import (
	"testing"

	"github.com/sed-evaluacion-desempeno/api/internal/pkg/quadrant"
	"github.com/stretchr/testify/assert"
)

// TestComputeQuadrant_AllCombinations covers every (perf, pot) pair from 1–9.
// Quadrant numbering (potential tier as rows, performance tier as columns):
//
//	                  Performance
//	                Low(1-3) Med(4-6) High(7-9)
//	Pot High(7-9)  |   7   |   8   |    9   |
//	Pot Med(4-6)   |   4   |   5   |    6   |
//	Pot Low(1-3)   |   1   |   2   |    3   |
func TestComputeQuadrant_AllCombinations(t *testing.T) {
	tests := []struct {
		perfRange [2]int // inclusive
		potRange  [2]int // inclusive
		want      int
	}{
		{[2]int{1, 3}, [2]int{1, 3}, 1},
		{[2]int{4, 6}, [2]int{1, 3}, 2},
		{[2]int{7, 9}, [2]int{1, 3}, 3},
		{[2]int{1, 3}, [2]int{4, 6}, 4},
		{[2]int{4, 6}, [2]int{4, 6}, 5},
		{[2]int{7, 9}, [2]int{4, 6}, 6},
		{[2]int{1, 3}, [2]int{7, 9}, 7},
		{[2]int{4, 6}, [2]int{7, 9}, 8},
		{[2]int{7, 9}, [2]int{7, 9}, 9},
	}

	for _, tt := range tests {
		for perf := tt.perfRange[0]; perf <= tt.perfRange[1]; perf++ {
			for pot := tt.potRange[0]; pot <= tt.potRange[1]; pot++ {
				t.Run("", func(t *testing.T) {
					got := quadrant.ComputeQuadrant(perf, pot)
					assert.Equalf(t, tt.want, got,
						"ComputeQuadrant(perf=%d, pot=%d) = %d; want %d",
						perf, pot, got, tt.want)
				})
			}
		}
	}
}

func TestComputeQuadrant_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		perf int
		pot  int
	}{
		{"zero performance", 0, 5},
		{"zero potential", 5, 0},
		{"both zero", 0, 0},
		{"performance 10", 10, 5},
		{"potential 10", 5, 10},
		{"both 10", 10, 10},
		{"negative performance", -1, 5},
		{"negative potential", 5, -1},
		{"both negative", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quadrant.ComputeQuadrant(tt.perf, tt.pot)
			assert.Equal(t, 0, got, "expected 0 for out-of-range input")
		})
	}
}

func TestComputeQuadrant_Corners(t *testing.T) {
	tests := []struct {
		name string
		perf int
		pot  int
		want int
	}{
		{"bottom-left (1,1)", 1, 1, 1},
		{"top-left (1,9)", 1, 9, 7},
		{"bottom-right (9,1)", 9, 1, 3},
		{"top-right (9,9)", 9, 9, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quadrant.ComputeQuadrant(tt.perf, tt.pot)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestComputeQuadrant_Middle(t *testing.T) {
	got := quadrant.ComputeQuadrant(5, 5)
	assert.Equal(t, 5, got, "(5,5) should map to quadrant 5")
}

func TestComputeQuadrant_Deterministic(t *testing.T) {
	const iterations = 1000
	for i := 0; i < iterations; i++ {
		got := quadrant.ComputeQuadrant(7, 3)
		assert.Equal(t, 3, got, "same input must always produce same output")
	}
}
