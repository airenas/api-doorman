package utils

import "math"

const float64EqualityThreshold float64 = 1e-6

// Float64Equal tests if two numbers are equal
func Float64Equal(f1, f2 float64) bool {
	return math.Abs(f1-f2) < float64EqualityThreshold
}
