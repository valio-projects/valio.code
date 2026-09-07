package ai

import (
	"errors"
	"math"
)

func vector(value []float32, expected int) ([]float32, error) {
	if len(value) != expected || expected < 1 {
		return nil, errors.New("AI response dimension mismatch")
	}
	var norm float64
	for _, coordinate := range value {
		if math.IsNaN(float64(coordinate)) || math.IsInf(float64(coordinate), 0) {
			return nil, errors.New("AI response is non-finite")
		}
		norm += float64(coordinate) * float64(coordinate)
	}
	if norm == 0 {
		return nil, errors.New("AI response is zero vector")
	}
	return value, nil
}
