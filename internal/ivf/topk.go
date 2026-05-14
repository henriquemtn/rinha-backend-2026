package ivf

import (
	"math"

	"rinha-backend-2026-go/internal/vector"
)

const maxProbe = 28

func pickTopFromDists(distances []float64, K, nProbe int, chosen *[maxProbe]uint32) []uint32 {
	var chosenDistances [28]float64
	worst := math.MaxFloat64
	worstIdx := 0
	count := 0

	for c := 0; c < K; c++ {
		d := distances[c]
		if count < nProbe {
			chosen[count] = uint32(c)
			chosenDistances[count] = d
			count++
			if count == nProbe {
				worstIdx = indexOfMax(chosenDistances[:nProbe])
				worst = chosenDistances[worstIdx]
			}
			continue
		}
		if d < worst {
			chosen[worstIdx] = uint32(c)
			chosenDistances[worstIdx] = d
			worstIdx = indexOfMax(chosenDistances[:nProbe])
			worst = chosenDistances[worstIdx]
		}
	}
	sortChosen(chosen, &chosenDistances, count)
	return chosen[:count]
}

func sortChosen(chosen *[maxProbe]uint32, distances *[maxProbe]float64, count int) {
	for i := 1; i < count; i++ {
		c := chosen[i]
		d := distances[i]
		j := i - 1
		for j >= 0 && distances[j] > d {
			chosen[j+1] = chosen[j]
			distances[j+1] = distances[j]
			j--
		}
		chosen[j+1] = c
		distances[j+1] = d
	}
}

func indexOfMax(xs []float64) int {
	idx := 0
	for i := 1; i < len(xs); i++ {
		if xs[i] > xs[idx] {
			idx = i
		}
	}
	return idx
}

func updateTopK(topDistances *[vector.TopK]int64, topLabels *[vector.TopK]uint8, worstIdx int, candidateDist int64, candidateLabel uint8) int {
	if candidateDist >= topDistances[worstIdx] {
		return worstIdx
	}
	topDistances[worstIdx] = candidateDist
	topLabels[worstIdx] = candidateLabel
	newWorst := 0
	for k := 1; k < vector.TopK; k++ {
		if topDistances[k] > topDistances[newWorst] {
			newWorst = k
		}
	}
	return newWorst
}
