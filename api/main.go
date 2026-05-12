package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	// Set GOMAXPROCS to 1 because each API container is limited to 0.45 CPU.
	runtime.GOMAXPROCS(1)

	normalizationPath := envOrDefault("NORMALIZATION_PATH", "resources/normalization.json")
	mccRiskPath := envOrDefault("MCC_RISK_PATH", "resources/mcc_risk.json")
	referencesPath := envOrDefault("REFERENCES_PATH", "resources/references.json.gz")
	startLoad := time.Now()
	LoadResources(normalizationPath, mccRiskPath, referencesPath)
	logStartupStats(time.Since(startLoad))

	listenAddr := envOrDefault("LISTEN_ADDR", ":9999")

	mux := http.NewServeMux()

	mux.HandleFunc("/ready", Ready)
	mux.HandleFunc("/fraud-score", FraudScore)

	server := &http.Server{Handler: mux}

	listener, err := createListener(listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Server running on %s", listenAddr)
	log.Fatal(server.Serve(listener))
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func logStartupStats(loadDuration time.Duration) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	log.Printf("Resources loaded in %s", loadDuration.Round(time.Millisecond))
	log.Printf("References: %d", len(ReferenceLabels))
	log.Printf("Memory: alloc=%dMB sys=%dMB", mem.Alloc/1024/1024, mem.Sys/1024/1024)
}

func createListener(addr string) (net.Listener, error) {
	if strings.HasPrefix(addr, "/") {
		if err := os.MkdirAll(filepath.Dir(addr), 0o755); err != nil {
			return nil, err
		}
		_ = os.Remove(addr)
		listener, err := net.Listen("unix", addr)
		if err != nil {
			return nil, err
		}
		if err := os.Chmod(addr, 0o666); err != nil {
			listener.Close()
			return nil, err
		}
		return listener, nil
	}
	return net.Listen("tcp", addr)
}
