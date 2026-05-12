package main

import (
	"errors"
	"math"
)

const (
	knnK           = 5
	fraudThreshold = 0.6
)

func CalculateScore(p Payload) (bool, float64, error) {
	if len(ReferenceVectors) == 0 {
		return true, 0, errors.New("reference vectors not loaded")
	}

	query, err := GenerateVector(p)
	if err != nil {
		return true, 0, err
	}

	var distances [knnK]float64
	var labels [knnK]string
	for i := 0; i < knnK; i++ {
		distances[i] = math.Inf(1)
	}

	for i := 0; i < len(ReferenceVectors); i++ {
		ref := &ReferenceVectors[i]
		dist := distanceSquared(query, ref.Vector)
		if dist >= distances[knnK-1] {
			continue
		}
		insertNeighbor(distances[:], labels[:], dist, ref.Label)
	}

	fraudCount := 0
	for i := 0; i < knnK; i++ {
		if labels[i] == "fraud" {
			fraudCount++
		}
	}

	fraudScore := float64(fraudCount) / float64(knnK)
	approved := fraudScore < fraudThreshold
	return approved, fraudScore, nil
}

func distanceSquared(a, b [14]float64) float64 {
	var sum float64
	for i := 0; i < 14; i++ {
		d := a[i] - b[i]
		sum += d * d
	}
	return sum
}

func insertNeighbor(distances []float64, labels []string, dist float64, label string) {
	pos := len(distances) - 1
	for pos > 0 && dist < distances[pos-1] {
		distances[pos] = distances[pos-1]
		labels[pos] = labels[pos-1]
		pos--
	}
	distances[pos] = dist
	labels[pos] = label
}
