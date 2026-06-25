package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"react-todos/apps/api/internal/delivery/http/dto"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
	Uptime    string            `json:"uptime"`
}

var startTime = time.Now()

// HealthCheckHandler returns a simple liveness response.
// For a full readiness check (including DB ping) use the /ready route in the router.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services: map[string]string{
			"api": "healthy",
		},
		Uptime: time.Since(startTime).String(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.SuccessResponse(health))
}
