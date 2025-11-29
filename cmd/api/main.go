package main

import (
	"context"
	"go-auth-service/internal/config"
	pLogger "go-auth-service/pkg/logger"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var (
		err error

		server *http.Server
	)

	pLogger.Info("Loading environment variables")
	envVariables, err := config.LoadServiceEnvironmentVariables()
	if err != nil {
		pLogger.Fatalf("Failed to load environment variables: %v", err)
	}
	pLogger.LogStruct("Environment variables loaded", envVariables)

	pLogger.Info("Initialize server configurations...")
	server = &http.Server{
		Addr:              envVariables.Server.Addr,
		ReadTimeout:       envVariables.Server.ReadTimeout,
		WriteTimeout:      envVariables.Server.WriteTimeout,
		IdleTimeout:       envVariables.Server.IdleTimeout,
		ReadHeaderTimeout: envVariables.Server.ReadHeaderTimeout,
	}
	pLogger.Info("Server configurations initialized")

	pLogger.Info("API server is running...")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			pLogger.Errorf("Failed to start server: %v", err)
		}

		return
	}()
	defer stop()
	<-ctx.Done()

	contextTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(contextTimeout); err != nil {
		pLogger.Fatalf("Failed to shutdown server gracefully: %v", err)
	}

	pLogger.Info("API server shutdown complete")
}
