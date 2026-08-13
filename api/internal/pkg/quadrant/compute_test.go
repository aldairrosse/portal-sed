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

// --- NEW: ComputePerformanceTier ---

func TestComputePerformanceTier_Boundaries(t *testing.T) {
	tests := []struct {
		name        string
		avgProgress float64
		want        int
	}{
		{"zero progress", 0, 1},
		{"lower boundary tier 1", 33, 1},
		{"upper boundary tier 1", 33.999, 1},
		{"lower boundary tier 2", 34, 2},
		{"middle tier 2", 50, 2},
		{"upper boundary tier 2", 66, 2},
		{"upper boundary tier 2 fractional", 66.999, 2},
		{"lower boundary tier 3", 67, 3},
		{"full progress", 100, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quadrant.ComputePerformanceTier(tt.avgProgress)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestComputePerformanceTier_NoData(t *testing.T) {
	// When avgProgress is 0 (no measurable progress yet), it's still tier 1
	got := quadrant.ComputePerformanceTier(0)
	assert.Equal(t, 1, got)
}

func TestComputePerformanceTier_Negative(t *testing.T) {
	// Negative progress isn't expected in practice, but treat as tier 1
	got := quadrant.ComputePerformanceTier(-5)
	assert.Equal(t, 1, got)
}

func TestComputePerformanceTier_ExactBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  int
	}{
		{"exact 33%", 33, 1},
		{"exact 34%", 34, 2},
		{"exact 66%", 66, 2},
		{"exact 67%", 67, 3},
		{"exact 100%", 100, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quadrant.ComputePerformanceTier(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- NEW: ComputePotentialTier ---

func TestComputePotentialTier_BothProvided(t *testing.T) {
	self := float64(4)
	hr := float64(4)
	got := quadrant.ComputePotentialTier(&self, &hr)
	assert.Equal(t, 3, got, "avg 4.0 → tier 3")
}

func TestComputePotentialTier_Boundaries(t *testing.T) {
	tests := []struct {
		name      string
		self      *float64
		hr        *float64
		want      int
	}{
		{"both 1.0 → tier 1", ptr(1.0), ptr(1.0), 1},
		{"both 2.33 → tier 1", ptr(2.33), ptr(2.33), 1},
		{"both 2.34 → tier 2", ptr(2.34), ptr(2.34), 2},
		{"both 3.66 → tier 2", ptr(3.66), ptr(3.66), 2},
		{"both 3.67 → tier 3", ptr(3.67), ptr(3.67), 3},
		{"both 5.0 → tier 3", ptr(5.0), ptr(5.0), 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quadrant.ComputePotentialTier(tt.self, tt.hr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestComputePotentialTier_OnlySelf(t *testing.T) {
	self := float64(2.0)
	got := quadrant.ComputePotentialTier(&self, nil)
	assert.Equal(t, 1, got, "self=2.0 alone → tier 1")
}

func TestComputePotentialTier_OnlyHR(t *testing.T) {
	hr := float64(4.5)
	got := quadrant.ComputePotentialTier(nil, &hr)
	assert.Equal(t, 3, got, "hr=4.5 alone → tier 3")
}

func TestComputePotentialTier_NoData(t *testing.T) {
	got := quadrant.ComputePotentialTier(nil, nil)
	assert.Equal(t, 2, got, "no data → default tier 2")
}

func TestComputePotentialTier_AverageRounding(t *testing.T) {
	self := float64(3.0)
	hr := float64(4.0)
	// avg = 3.5 → tier 2 (2.34-3.66)
	got := quadrant.ComputePotentialTier(&self, &hr)
	assert.Equal(t, 2, got)
}

func TestComputePotentialTier_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		self *float64
		hr   *float64
		want int
	}{
		{"self 1.0, hr 5.0 avg 3.0 → tier 2", ptr(1.0), ptr(5.0), 2},
		{"self 1.0 only → tier 1", ptr(1.0), nil, 1},
		{"hr 5.0 only → tier 3", nil, ptr(5.0), 3},
		{"self 2.33, hr 2.34 avg 2.335 → tier 2", ptr(2.33), ptr(2.34), 2},
		{"exact 2.34 avg → tier 2", ptr(2.34), ptr(2.34), 2},
		{"exact 3.66 avg → tier 2", ptr(3.66), ptr(3.66), 2},
		{"exact 3.67 avg → tier 3", ptr(3.67), ptr(3.67), 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quadrant.ComputePotentialTier(tt.self, tt.hr)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- NEW: ComputeWeightedPotentialTier ---

func TestComputeWeightedPotentialTier_RHWeighted(t *testing.T) {
	// 5*0.2 + 3*0.8 = 1.0 + 2.4 = 3.4 → tier 2 (≤ 3.66)
	got := quadrant.ComputeWeightedPotentialTier(5, 3, 0.2, 0.8)
	assert.Equal(t, 2, got, "weighted avg 3.4 → tier 2")
}

// --- NEW: ComputeQuadrantFromTiers ---

func TestComputeQuadrantFromTiers_All9Combinations(t *testing.T) {
	// Quadrant = (potTier-1)*3 + perfTier
	tests := []struct {
		perfTier int
		potTier  int
		want     int
	}{
		{1, 1, 1},
		{2, 1, 2},
		{3, 1, 3},
		{1, 2, 4},
		{2, 2, 5},
		{3, 2, 6},
		{1, 3, 7},
		{2, 3, 8},
		{3, 3, 9},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := quadrant.ComputeQuadrantFromTiers(tt.perfTier, tt.potTier)
			assert.Equal(t, tt.want, got,
				"ComputeQuadrantFromTiers(perfTier=%d, potTier=%d) = %d; want %d",
				tt.perfTier, tt.potTier, got, tt.want)
		})
	}
}

func TestComputeQuadrantFromTiers_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		perfTier int
		potTier  int
	}{
		{"perf zero", 0, 2},
		{"pot zero", 2, 0},
		{"both zero", 0, 0},
		{"perf 4 (out of range)", 4, 2},
		{"pot 4 (out of range)", 2, 4},
		{"both negative", -1, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quadrant.ComputeQuadrantFromTiers(tt.perfTier, tt.potTier)
			assert.Equal(t, 0, got, "expected 0 for out-of-range input")
		})
	}
}

func TestComputeQuadrantFromTiers_Midpoint(t *testing.T) {
	got := quadrant.ComputeQuadrantFromTiers(2, 2)
	assert.Equal(t, 5, got, "(2,2) → quadrant 5")
}

func TestComputeQuadrantFromTiers_Corners(t *testing.T) {
	tests := []struct {
		name     string
		perfTier int
		potTier  int
		want     int
	}{
		{"bottom-left (1,1)", 1, 1, 1},
		{"top-left (1,3)", 1, 3, 7},
		{"bottom-right (3,1)", 3, 1, 3},
		{"top-right (3,3)", 3, 3, 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quadrant.ComputeQuadrantFromTiers(tt.perfTier, tt.potTier)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ptr is a helper for *float64.
func ptr(v float64) *float64 {
	return &v
}
