package services

import "math"

func normalize(vec []float64) []float64 {
	var norm float64

	for _, v := range vec {
		norm += v * v
	}

	if norm == 0 {
		return vec
	}

	norm = math.Sqrt(norm)

	normalized := make([]float64, len(vec))
	for i, v := range vec {
		normalized[i] = v / norm
	}

	return normalized
}

func dotProduct(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dot float64
	for i := range a {
		dot += a[i] * b[i]
	}

	return dot
}
