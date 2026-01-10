package handler

import (
	"go-auth-service/pkg/response"
	"net/http"
	"time"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	data := HealthResponse{
		Status:    "up",
		Timestamp: time.Now(),
		Service:   "go-auth-service",
	}

	response.Success(w, data, `Service is healthy`)
}
