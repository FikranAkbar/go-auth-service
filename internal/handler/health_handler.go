package handler

import (
	"go-auth-service/pkg/constants"
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

const (
	healthStatusUp = "up"
	healthMessage  = "Service is healthy"
)

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	data := HealthResponse{
		Status:    healthStatusUp,
		Timestamp: time.Now(),
		Service:   constants.AppName,
	}

	response.Success(w, data, healthMessage)
}
