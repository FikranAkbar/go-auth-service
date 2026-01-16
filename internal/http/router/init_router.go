package router

import (
	"go-auth-service/internal/config"
	domainSecurity "go-auth-service/internal/domain/security"
	"go-auth-service/internal/handler"
	authMiddleware "go-auth-service/internal/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitRouter(envs *config.Env, jwtManager domainSecurity.JWTManagerInterface, handlers ...interface{}) *chi.Mux {
	var (
		userHandler   = handlers[0].(*handler.UserHandler)
		healthHandler = handlers[1].(*handler.HealthHandler)
		authHandler   = handlers[2].(*handler.AuthHandler)
	)

	// Initialize auth middleware
	authMW := authMiddleware.NewAuthMiddleware(jwtManager)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(envs.Server.WriteTimeout))

	// Health check endpoint
	r.Get(`/health`, healthHandler.HealthCheck)

	// Auth endpoints (public)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Get("/verify-email", authHandler.VerifyEmail)
		r.Post("/resend-verification", authHandler.ResendVerificationEmail)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh-token", authHandler.RefreshToken)

		// Protected endpoint (requires authentication)
		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Post("/logout", authHandler.Logout)
		})
	})

	// User endpoints
	r.Route("/api/users", func(r chi.Router) {
		// Public endpoint to get user by ID (returns public user info only)
		r.Get("/{id}", userHandler.GetUserByID)

		// Protected endpoints (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Get("/me", userHandler.GetCurrentUser)
		})
	})

	return r
}
