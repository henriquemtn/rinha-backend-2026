package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
)

type Normalization struct {
	MaxAmount            float64 `json:"max_amount"`
	MaxInstallments      float64 `json:"max_installments"`
	AmountVsAvgRatio     float64 `json:"amount_vs_avg_ratio"`
	MaxMinutes           float64 `json:"max_minutes"`
	MaxKm                float64 `json:"max_km"`
	MaxTxCount24h        float64 `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount float64 `json:"max_merchant_avg_amount"`
}

type referenceDTO struct {
	Vector [14]float32 `json:"vector"`
	Label  string      `json:"label"`
}

const vectorDims = 14

var NormalizationData Normalization
var ReferenceVectors []uint16
var ReferenceLabels []bool
var MccRisk map[string]float64

func LoadNormalization(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Erro ao abrir normalization.json: %v", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&NormalizationData); err != nil {
		log.Fatalf("Erro ao decodificar normalization.json: %v", err)
	}
}

func LoadReferences(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Erro ao abrir references.json.gz: %v", err)
	}
	defer f.Close()
	reader, closer, err := openMaybeGzip(f)
	if err != nil {
		log.Fatalf("Erro ao abrir referencias: %v", err)
	}
	if closer != nil {
		defer closer.Close()
	}
	dec := json.NewDecoder(reader)
	if token, err := dec.Token(); err != nil {
		log.Fatalf("Erro ao ler início do array: %v", err)
	} else if delim, ok := token.(json.Delim); !ok || delim != '[' {
		log.Fatalf("Formato inválido em references: esperado '['")
	}
	for dec.More() {
		var ref referenceDTO
		if err := dec.Decode(&ref); err != nil {
			log.Fatalf("Erro ao decodificar referência: %v", err)
		}
		for i := 0; i < vectorDims; i++ {
			ReferenceVectors = append(ReferenceVectors, quantizeVectorValue(ref.Vector[i]))
		}
		ReferenceLabels = append(ReferenceLabels, ref.Label == "fraud")
	}
}

func quantizeVectorValue(v float32) uint16 {
	if v < -1 {
		v = -1
	}
	if v > 1 {
		v = 1
	}
	return uint16((v + 1) * 32767.5)
}

func openMaybeGzip(r io.Reader) (io.Reader, io.Closer, error) {
	buf := bufio.NewReader(r)
	peek, err := buf.Peek(2)
	if err == nil && len(peek) == 2 && peek[0] == 0x1f && peek[1] == 0x8b {
		gz, err := gzip.NewReader(buf)
		if err != nil {
			return nil, nil, err
		}
		return gz, gz, nil
	}
	return buf, nil, nil
}

func LoadMccRisk(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Erro ao abrir mcc_risk.json: %v", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&MccRisk); err != nil {
		log.Fatalf("Erro ao decodificar mcc_risk.json: %v", err)
	}
}

func LoadResources(normalizationPath, mccRiskPath, referencesPath string) {
	LoadNormalization(normalizationPath)
	LoadMccRisk(mccRiskPath)
	LoadReferences(referencesPath)

	// Build IVF index after references are loaded to speed up queries.
	k, samples, iterations, nProbe := loadIVFConfig()
	if k > 0 {
		if err := sanityCheckIVFConfig(k, nProbe); err != nil {
			log.Printf("IVF disabled: %v", err)
			return
		}
		if err := BuildIVFIndex(k, samples, iterations, nProbe); err != nil {
			log.Printf("IVF disabled: %v", err)
		}
	}
}

func loadIVFConfig() (int, int, int, int) {
	// Defaults tuned for faster startup with measurable speedups.
	const (
		defaultK          = 1024
		defaultSamples    = 50000
		defaultIterations = 8
		defaultNProbe     = 8
	)

	return envIntOrDefault("IVF_K", defaultK),
		envIntOrDefault("IVF_TRAIN_SAMPLES", defaultSamples),
		envIntOrDefault("IVF_ITER", defaultIterations),
		envIntOrDefault("N_PROBE", defaultNProbe)
}

func envIntOrDefault(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
