package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateTraceID generates a unique trace ID
// Format: 32 character hexadecimal string (similar to UUID without dashes)
func GenerateTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// ContextWithTraceID creates a new context with a trace ID
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// ContextWithNewTraceID creates a new context with a newly generated trace ID
func ContextWithNewTraceID(ctx context.Context) context.Context {
	return ContextWithTraceID(ctx, GenerateTraceID())
}

// GetTraceIDFromContext extracts trace ID from context
func GetTraceIDFromContext(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(TraceIDKey).(string)
	return traceID, ok
}

// ContextWithRequestID creates a new context with a request ID
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}

// ContextWithUserID creates a new context with a user ID
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// ContextWithIDs creates a new context with trace ID, request ID, and optionally user ID
func ContextWithIDs(ctx context.Context, traceID, requestID, userID string) context.Context {
	ctx = ContextWithTraceID(ctx, traceID)
	ctx = ContextWithRequestID(ctx, requestID)
	if userID != "" {
		ctx = ContextWithUserID(ctx, userID)
	}
	return ctx
}
