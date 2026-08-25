package main

import (
	"encoding/json"
	"net/http"
)

func decodeCreate(r *http.Request) (CreatePassRequest, error) {
	var req CreatePassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, ErrInvalidPass
	}
	return req, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
