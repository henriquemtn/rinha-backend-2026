package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func Ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func FraudScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload Payload

	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	approved, score := CalculateScore(payload)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"approved":    approved,
		"fraud_score": score,
	})
}
