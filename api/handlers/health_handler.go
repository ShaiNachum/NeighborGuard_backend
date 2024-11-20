package handlers

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status string
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	newHealthResponse := HealthResponse{Status: "healthy"}
	json.NewEncoder(w).Encode(newHealthResponse)
}
