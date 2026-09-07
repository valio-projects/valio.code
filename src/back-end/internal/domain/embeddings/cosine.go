package embeddings

import (
	"errors"
	"math"
)

// Cosine returns exact similarity for two finite, nonzero vectors of equal length.
// It performs no approximate-nearest-neighbor indexing.
func Cosine(a, b []float32) (float32, error) {
	if len(a) == 0 || len(a) != len(b) || !finiteNonZero(a) || !finiteNonZero(b) {
		return 0, errors.New("invalid cosine vectors")
	}
	var dot, an, bn float64
	for index := range a {
		dot += float64(a[index]) * float64(b[index])
		an += float64(a[index]) * float64(a[index])
		bn += float64(b[index]) * float64(b[index])
	}
	return float32(dot / math.Sqrt(an*bn)), nil
}

func finiteNonZero(values []float32) bool {
	var norm float64
	for _, value := range values {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return false
		}
		norm += float64(value) * float64(value)
	}
	return norm > 0 && !math.IsInf(norm, 0) && !math.IsNaN(norm)
}
