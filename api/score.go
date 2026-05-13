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

	// Prefer IVF search when available to reduce candidates.
	if IVFEnabled {
		return scoreWithIVF(query)
	}

	return scoreWithFullScan(query)
}

func scoreWithFullScan(query [14]float64) (bool, float64, error) {
	// Baseline: full scan over all references.
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

	return finalizeScore(labels)
}

func scoreWithIVF(query [14]float64) (bool, float64, error) {
	// IVF: probe a subset of centroid lists before scoring.
	var distances [knnK]float64
	var labels [knnK]bool
	for i := 0; i < knnK; i++ {
		distances[i] = math.Inf(1)
	}

	centroidIndexes := selectNearestCentroids(query)
	for _, centroidIndex := range centroidIndexes {
		list := IVFLists[centroidIndex]
		for _, refIndex := range list {
			dist := distanceSquared(query, refIndex)
			if dist >= distances[knnK-1] {
				continue
			}
			insertNeighbor(distances[:], labels[:], dist, ReferenceLabels[refIndex])
		}
	}

	return finalizeScore(labels)
}

func finalizeScore(labels [knnK]bool) (bool, float64, error) {
	// Convert top-k labels into final fraud score.
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
