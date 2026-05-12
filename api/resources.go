package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"os"
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

type Reference struct {
	Vector [14]float64 `json:"vector"`
	Label  string      `json:"label"`
}

var NormalizationData Normalization
var ReferenceVectors []Reference
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
		var ref Reference
		if err := dec.Decode(&ref); err != nil {
			log.Fatalf("Erro ao decodificar referência: %v", err)
		}
		ReferenceVectors = append(ReferenceVectors, ref)
	}
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
}
