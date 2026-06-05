// Package validation provides weight validation helpers for the Double 100% rule.
package validation

import "math"

// Epsilon is the tolerance for floating-point weight comparisons (0.01%).
const Epsilon = 0.01

// WithinEpsilon checks whether a and b are within Epsilon of each other.
func WithinEpsilon(a, b float64) bool {
	return math.Abs(a-b) <= Epsilon
}

// SumValid sums a slice of float64 values and reports whether the sum
// is within Epsilon of expected. For a nil or empty slice, returns 0 and
// valid=true only if expected is 0 or within Epsilon of 0.
func SumValid(items []float64, expected float64) (sum float64, valid bool) {
	for _, v := range items {
		sum += v
	}
	return sum, WithinEpsilon(sum, expected)
}

// WeightedSumFromMap sums all weight values in a map[string]float64.
// Useful for summing weights across categories or goals keyed by ID.
func WeightedSumFromMap(weights map[string]float64) float64 {
	var sum float64
	for _, w := range weights {
		sum += w
	}
	return sum
}
