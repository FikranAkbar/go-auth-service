package main

import (
	"context"
	"database/sql"
	"go-auth-service/internal/config"
	"go-auth-service/internal/handler"
	"go-auth-service/internal/http/router"
	"go-auth-service/internal/infra"
	"go-auth-service/internal/repository"
	"go-auth-service/internal/security"
	"go-auth-service/internal/service"
	pLogger "go-auth-service/pkg/logger"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

// AppDependencies holds all external dependencies needed by the application
type AppDependencies struct {
	Config      *config.Env
	DB          *sql.DB
	RedisClient *redis.Client
}
type App struct {
	HttpServer *http.Server
}

// NewApp wires up all application components.
// HTTP handlers will use r.Context() from requests for request-scoped operations.
func NewApp(deps AppDependencies) *App {
	var (
		app            = &App{}
		userRepository *repository.UserRepository
		tokenRepo      *repository.TokenRepository
		userService    *service.UserService
		emailService   *service.EmailService
		authService    *service.AuthService
		userHandler    *handler.UserHandler
		authHandler    *handler.AuthHandler
	)
	pLogger.Info("Creating handlers")

	// Initialize security components
	passwordHasher := security.NewPasswordHasher(deps.Config.JWT.BcryptCost)
	jwtManager := security.NewJWTManager(&deps.Config.JWT)

	// Initialize email client and service
	emailClient := infra.NewEmailClient(&deps.Config.Email)
	emailService = service.NewEmailService(emailClient, deps.Config.App.URL)

	// Initialize repository and services
	userRepository = repository.NewUserRepository(deps.DB, deps.RedisClient)
	tokenRepo = repository.NewTokenRepository(deps.RedisClient)
	userService = service.NewUserService(userRepository, passwordHasher)
	authService = service.NewAuthService(userService, emailService, jwtManager, tokenRepo)

	// Initialize handlers
	userHandler = handler.NewUserHandler(userService)
	authHandler = handler.NewAuthHandler(authService, jwtManager)
	healthHandler := handler.NewHealthHandler()

	pLogger.Info("Handlers created")
	pLogger.Info("Initialize server configurations...")
	app.HttpServer = &http.Server{
		Handler:           router.InitRouter(deps.Config, jwtManager, userHandler, healthHandler, authHandler),
		Addr:              deps.Config.Server.Port,
		ReadTimeout:       deps.Config.Server.ReadTimeout,
		WriteTimeout:      deps.Config.Server.WriteTimeout,
		IdleTimeout:       deps.Config.Server.IdleTimeout,
		ReadHeaderTimeout: deps.Config.Server.ReadHeaderTimeout,
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
	// Application lifecycle context (for infrastructure setup only)
	ctx := context.Background()
	// Load configuration
	pLogger.Info("Loading environment variables")
	envVariables := config.LoadServiceEnvironmentVariables()
	pLogger.LogStruct("Environment variables loaded", envVariables)
	// Setup infrastructure
	pLogger.Info("Connecting to Postgres database")
	postgresDB := infra.NewPostgresDB(envVariables)
	pLogger.Info("Connected to Postgres database")
	pLogger.Info("Connecting to Redis database")
	redisClient := infra.NewRedisClient(ctx, *envVariables)
	pLogger.Info("Connected to Redis database")
	// Create dependencies
	deps := AppDependencies{
		Config:      envVariables,
		DB:          postgresDB,
		RedisClient: redisClient,
	}
	// Run app
	NewApp(deps).RunApp()
}
