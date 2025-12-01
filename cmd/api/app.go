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

type App struct {
	HttpServer *http.Server
}

func NewApp(context context.Context) *App {
	var (
		err error
		app = &App{}

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
	app.HttpServer = &http.Server{
		Handler:           router.InitRouter(envVariables, userHandler, authHandler),
		Addr:              envVariables.Server.Port,
		ReadTimeout:       envVariables.Server.ReadTimeout,
		WriteTimeout:      envVariables.Server.WriteTimeout,
		IdleTimeout:       envVariables.Server.IdleTimeout,
		ReadHeaderTimeout: envVariables.Server.ReadHeaderTimeout,
	}
	pLogger.Info("Server configurations initialized")

	return app
}

func (app *App) RunApp() {
	var (
		server = app.HttpServer
	)

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
