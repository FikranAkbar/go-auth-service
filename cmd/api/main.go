package main

import (
	"context"
	"go-auth-service/internal/config"
	"go-auth-service/internal/handler"
	"go-auth-service/internal/http/router"
	pLogger "go-auth-service/pkg/logger"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var (
		err error

		server      *http.Server
		userHandler *handler.UserHandler
		authHandler *handler.AuthHandler
	)

	pLogger.Info("Loading environment variables")
	envVariables, err := config.LoadServiceEnvironmentVariables()
	if err != nil {
		pLogger.Fatalf("Failed to load environment variables: %v", err)
	}
	pLogger.LogStruct("Environment variables loaded", envVariables)

	pLogger.Info("Creating handlers")
	userHandler = handler.NewUserHandler()
	authHandler = handler.NewAuthHandler()
	pLogger.Info("Handlers created")

	pLogger.Info("Initialize server configurations...")
	server = &http.Server{
		Handler:           router.InitRouter(envVariables, userHandler, authHandler),
		Addr:              envVariables.Server.Port,
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
			return
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
