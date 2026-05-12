package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
)

func main() {
	// Set GOMAXPROCS to 1 because each API container is limited to 0.45 CPU.
	runtime.GOMAXPROCS(1)

	normalizationPath := envOrDefault("NORMALIZATION_PATH", "resources/normalization.json")
	mccRiskPath := envOrDefault("MCC_RISK_PATH", "resources/mcc_risk.json")
	referencesPath := envOrDefault("REFERENCES_PATH", "resources/references.json.gz")
	LoadResources(normalizationPath, mccRiskPath, referencesPath)

	mux := http.NewServeMux()

	mux.HandleFunc("/ready", Ready)
	mux.HandleFunc("/fraud-score", FraudScore)

	server := &http.Server{
		Addr:    ":9999",
		Handler: mux,
	}

	log.Println("Server running on :9999")
	log.Fatal(server.ListenAndServe())
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
