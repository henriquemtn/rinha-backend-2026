package main

import (
	"log"
	"net/http"
	"runtime"
)

func main() {
	// Set GOMAXPROCS to 1 because each API container is limited to 0.45 CPU.
	runtime.GOMAXPROCS(1)

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
