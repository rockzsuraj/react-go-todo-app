package handlers

import (
	"encoding/json"
	"net/http"
	"react-todos/apps/api/internal/dto"
	"time"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
	Uptime    string            `json:"uptime"`
}

var startTime = time.Now()

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Check database connection
	dbStatus := "healthy"
	if todoService == nil {
		dbStatus = "unhealthy"
	}
	
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services: map[string]string{
			"database": dbStatus,
			"api":      "healthy",
		},
		Uptime: time.Since(startTime).String(),
	}
	
	// If any service is unhealthy, mark overall status as unhealthy
	for _, status := range health.Services {
		if status != "healthy" {
			health.Status = "unhealthy"
			w.WriteHeader(http.StatusServiceUnavailable)
			break
		}
	}
	
	if health.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	}
	
	json.NewEncoder(w).Encode(dto.SuccessResponse(health))
}