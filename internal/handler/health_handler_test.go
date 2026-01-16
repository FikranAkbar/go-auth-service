package handler

import (
	"encoding/json"
	"go-auth-service/pkg/constants"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// HEALTH CHECK ENDPOINT TESTS
// Contract: GET /health or GET /api/health
// Input: None
// Success: Always returns 200 with service status
// Failures: None (this endpoint should always succeed)
// ============================================================================

func TestHealthHandler_HealthCheck_Success(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Service is healthy", response["message"])

	// Verify data
	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "up", data["status"])
	assert.Equal(t, constants.AppName, data["service"])

	// Verify timestamp exists and is valid
	timestamp, ok := data["timestamp"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, timestamp)

	// Parse timestamp to ensure it's valid
	parsedTime, err := time.Parse(time.RFC3339Nano, timestamp)
	assert.NoError(t, err)

	// Timestamp should be recent (within last 5 seconds)
	timeDiff := time.Since(parsedTime)
	assert.Less(t, timeDiff, 5*time.Second)
}

func TestHealthHandler_HealthCheck_ResponseFormat(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	// Verify Content-Type is JSON
	contentType := w.Header().Get("Content-Type")
	assert.Contains(t, contentType, "application/json")

	// Verify status code
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthHandler_HealthCheck_MultipleRequests(t *testing.T) {
	// Health check should work consistently across multiple requests
	handler := NewHealthHandler()

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthCheck(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, true, response["success"])
	}
}

func TestHealthHandler_HealthCheck_NilRequest(t *testing.T) {
	// Health check should handle nil request (underscore parameter)
	handler := NewHealthHandler()

	w := httptest.NewRecorder()

	// Pass nil as request (the handler uses _ for request parameter)
	handler.HealthCheck(w, nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthHandler_NewHealthHandler(t *testing.T) {
	// Test handler creation
	handler := NewHealthHandler()

	assert.NotNil(t, handler)
	assert.IsType(t, &HealthHandler{}, handler)
}
