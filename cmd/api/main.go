package main

import (
	"context"
	"database/sql"
	"go-auth-service/internal/config"
	"go-auth-service/internal/handler"
	"go-auth-service/internal/http/router"
	"go-auth-service/internal/infra"
	"go-auth-service/internal/repository"
	"go-auth-service/internal/service"
	pLogger "go-auth-service/pkg/logger"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	HttpServer *http.Server
}

func NewApp(context context.Context) *App {
	var (
		app = &App{}

		postgresDB  *sql.DB
		redisClient *redis.Client

		userRepository *repository.UserRepository
		userService    *service.UserService
		userHandler    *handler.UserHandler
	)

	pLogger.Info("Loading environment variables")
	envVariables := config.LoadServiceEnvironmentVariables()
	pLogger.LogStruct("Environment variables loaded", envVariables)

	pLogger.Info("Connecting to Postgres database")
	postgresDB = infra.NewPostgresDB(envVariables)
	pLogger.Info("Connected to Postgres database")

	pLogger.Info("Connecting to Redis database")
	redisClient = infra.NewRedisClient(context, *envVariables)
	pLogger.Info("Connected to Redis database")

	pLogger.Info("Creating handlers")
	userRepository = repository.NewUserRepository(postgresDB, redisClient)
	userService = service.NewUserService(userRepository)
	userHandler = handler.NewUserHandler(userService)

	healthHandler := handler.NewHealthHandler()
	pLogger.Info("Handlers created")

	pLogger.Info("Initialize server configurations...")
	app.HttpServer = &http.Server{
		Handler:           router.InitRouter(envVariables, userHandler, healthHandler),
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
		pLogger.Errorf("Failed to shutdown server gracefully: %v", err)
	}

	pLogger.Info("API server shutdown complete")
}

func main() {
	ctx := context.Background()
	NewApp(ctx).RunApp()
}
