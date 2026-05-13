package main

import (
	"errors"
	"fmt"
	"log"
	"math"
)

type Vector14 [vectorDims]float32

type Vector14F [vectorDims]float64

var (
	IVFEnabled   bool
	IVFProbe     int
	IVFCentroids []Vector14
	IVFLists     [][]int
)

func BuildIVFIndex(k, trainSamples, iterations, nProbe int) error {
	// Validate configuration and input data.
	if len(ReferenceLabels) == 0 {
		return errors.New("references not loaded")
	}
	if k <= 0 {
		return errors.New("IVF_K must be > 0")
	}
	if k > len(ReferenceLabels) {
		k = len(ReferenceLabels)
	}
	if trainSamples <= 0 {
		trainSamples = len(ReferenceLabels)
	}
	if trainSamples > len(ReferenceLabels) {
		trainSamples = len(ReferenceLabels)
	}
	if iterations <= 0 {
		iterations = 8
	}
	if nProbe <= 0 {
		nProbe = 8
	}

	// Sample a subset of references for k-means training.
	samples := make([]Vector14, trainSamples)
	stride := len(ReferenceLabels) / trainSamples
	if stride < 1 {
		stride = 1
	}
	for i := 0; i < trainSamples; i++ {
		refIndex := i * stride
		if refIndex >= len(ReferenceLabels) {
			refIndex = len(ReferenceLabels) - 1
		}
		samples[i] = loadReferenceVector(refIndex)
	}

	// Initialize centroids using evenly spaced samples.
	centroids := make([]Vector14, k)
	centroidStride := trainSamples / k
	if centroidStride < 1 {
		centroidStride = 1
	}
	for i := 0; i < k; i++ {
		idx := i * centroidStride
		if idx >= trainSamples {
			idx = trainSamples - 1
		}
		centroids[i] = samples[idx]
	}

	// Run a lightweight k-means to refine centroids.
	for iter := 0; iter < iterations; iter++ {
		sums := make([]Vector14F, k)
		counts := make([]int, k)
		for _, sample := range samples {
			nearest := nearestCentroidForVector(sample, centroids)
			counts[nearest]++
			for d := 0; d < vectorDims; d++ {
				sums[nearest][d] += float64(sample[d])
			}
		}
		for i := 0; i < k; i++ {
			if counts[i] == 0 {
				continue
			}
			inv := 1.0 / float64(counts[i])
			for d := 0; d < vectorDims; d++ {
				centroids[i][d] = float32(sums[i][d] * inv)
			}
		}
	}

	// Assign each reference to its nearest centroid.
	assignments := make([]int, len(ReferenceLabels))
	counts := make([]int, k)
	for i := 0; i < len(ReferenceLabels); i++ {
		nearest := nearestCentroidForRef(i, centroids)
		assignments[i] = nearest
		counts[nearest]++
	}

	lists := make([][]int, k)
	for i := 0; i < k; i++ {
		lists[i] = make([]int, 0, counts[i])
	}
	for i, centroidIndex := range assignments {
		lists[centroidIndex] = append(lists[centroidIndex], i)
	}

	IVFCentroids = centroids
	IVFLists = lists
	IVFEnabled = true
	IVFProbe = nProbe

	log.Printf("IVF ready: k=%d samples=%d iterations=%d n_probe=%d", k, trainSamples, iterations, nProbe)
	return nil
}

func selectNearestCentroids(query [14]float64) []int {
	// If probes exceed centroids, just scan all of them.
	if IVFProbe >= len(IVFCentroids) {
		indexes := make([]int, len(IVFCentroids))
		for i := range IVFCentroids {
			indexes[i] = i
		}
		return indexes
	}

	distances := make([]float64, IVFProbe)
	indexes := make([]int, IVFProbe)
	for i := 0; i < IVFProbe; i++ {
		distances[i] = math.Inf(1)
		indexes[i] = -1
	}

	for i, centroid := range IVFCentroids {
		dist := distanceToCentroid(query, centroid)
		if dist >= distances[IVFProbe-1] {
			continue
		}
		insertCentroid(distances, indexes, dist, i)
	}

	for i := 0; i < len(indexes); i++ {
		if indexes[i] < 0 {
			return indexes[:i]
		}
	}
	return indexes
}

func insertCentroid(distances []float64, indexes []int, dist float64, index int) {
	pos := len(distances) - 1
	for pos > 0 && dist < distances[pos-1] {
		distances[pos] = distances[pos-1]
		indexes[pos] = indexes[pos-1]
		pos--
	}
	distances[pos] = dist
	indexes[pos] = index
}

func nearestCentroidForVector(v Vector14, centroids []Vector14) int {
	best := 0
	bestDist := math.Inf(1)
	for i, centroid := range centroids {
		dist := distanceVectorToCentroid(v, centroid)
		if dist < bestDist {
			bestDist = dist
			best = i
		}
	}
	return best
}

func nearestCentroidForRef(refIndex int, centroids []Vector14) int {
	best := 0
	bestDist := math.Inf(1)
	for i, centroid := range centroids {
		dist := distanceRefToCentroid(refIndex, centroid)
		if dist < bestDist {
			bestDist = dist
			best = i
		}
	}
	return best
}

func loadReferenceVector(refIndex int) Vector14 {
	var v Vector14
	base := refIndex * vectorDims
	for i := 0; i < vectorDims; i++ {
		v[i] = float32(dequantizeVectorValue(ReferenceVectors[base+i]))
	}
	return v
}

func distanceToCentroid(query [14]float64, centroid Vector14) float64 {
	var sum float64
	for i := 0; i < vectorDims; i++ {
		d := query[i] - float64(centroid[i])
		sum += d * d
	}
	return sum
}

func distanceVectorToCentroid(v Vector14, centroid Vector14) float64 {
	var sum float64
	for i := 0; i < vectorDims; i++ {
		d := float64(v[i] - centroid[i])
		sum += d * d
	}
	return sum
}

func distanceRefToCentroid(refIndex int, centroid Vector14) float64 {
	var sum float64
	base := refIndex * vectorDims
	for i := 0; i < vectorDims; i++ {
		value := dequantizeVectorValue(ReferenceVectors[base+i])
		d := value - float64(centroid[i])
		sum += d * d
	}
	return sum
}

func sanityCheckIVFConfig(k, nProbe int) error {
	if k <= 0 {
		return errors.New("IVF_K must be positive")
	}
	if nProbe <= 0 {
		return errors.New("N_PROBE must be positive")
	}
	if nProbe > k {
		return fmt.Errorf("N_PROBE (%d) must be <= IVF_K (%d)", nProbe, k)
	}
	return nil
}
