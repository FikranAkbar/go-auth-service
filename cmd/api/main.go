package main

import (
	"context"
	"go-auth-service/internal/config"
	pkgLogger "go-auth-service/pkg/logger"
	"log"
)

func main() {
	var (
		ctx    = pkgLogger.ContextWithTraceID(context.Background(), `service-startup`)
		logger = pkgLogger.WithContext(ctx)
	)

	// Load environment variables
	envVariables := must(config.LoadServiceEnvironmentVariables())
	logger.LogStruct("Environment variables loaded", envVariables)

	logger.Info("API server is running...")
}

func must[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}
