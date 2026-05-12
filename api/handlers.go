package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func Ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func FraudScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	var payload Payload

	if err := json.Unmarshal(body, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	approved, score, err := CalculateScore(payload)
	if err != nil {
		if errors.Is(err, ErrResourcesNotLoaded) {
			writeJSONError(w, http.StatusInternalServerError, "resources not loaded")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"approved":    approved,
		"fraud_score": score,
	})
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
