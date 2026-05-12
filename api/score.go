package main

import (
	"errors"
	"math"
)

const (
	knnK           = 5
	fraudThreshold = 0.6
)

var ErrResourcesNotLoaded = errors.New("resources not loaded")

func CalculateScore(p Payload) (bool, float64, error) {
	if len(ReferenceVectors) == 0 {
		return true, 0, ErrResourcesNotLoaded
	}

	query, err := GenerateVector(p)
	if err != nil {
		return true, 0, err
	}

	var distances [knnK]float64
	var labels [knnK]bool
	for i := 0; i < knnK; i++ {
		distances[i] = math.Inf(1)
	}

	for i := 0; i < len(ReferenceLabels); i++ {
		dist := distanceSquared(query, i)
		if dist >= distances[knnK-1] {
			continue
		}
		insertNeighbor(distances[:], labels[:], dist, ReferenceLabels[i])
	}

	fraudCount := 0
	for i := 0; i < knnK; i++ {
		if labels[i] {
			fraudCount++
		}
	}

	fraudScore := float64(fraudCount) / float64(knnK)
	approved := fraudScore < fraudThreshold
	return approved, fraudScore, nil
}

func distanceSquared(a [14]float64, refIndex int) float64 {
	var sum float64
	base := refIndex * vectorDims
	for i := 0; i < vectorDims; i++ {
		b := dequantizeVectorValue(ReferenceVectors[base+i])
		d := a[i] - b
		sum += d * d
	}
	return sum
}

func dequantizeVectorValue(v uint16) float64 {
	return float64(v)/32767.5 - 1
}

func insertNeighbor(distances []float64, labels []bool, dist float64, label bool) {
	pos := len(distances) - 1
	for pos > 0 && dist < distances[pos-1] {
		distances[pos] = distances[pos-1]
		labels[pos] = labels[pos-1]
		pos--
	}
	distances[pos] = dist
	labels[pos] = label
}
