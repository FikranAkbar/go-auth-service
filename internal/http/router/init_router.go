package router

import (
	"go-auth-service/internal/config"
	"go-auth-service/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitRouter(envs *config.Env, handlers ...interface{}) *chi.Mux {
	var (
		_             = handlers[0].(*handler.UserHandler)
		healthHandler = handlers[1].(*handler.HealthHandler)
		authHandler   = handlers[2].(*handler.AuthHandler)
	)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(envs.Server.WriteTimeout))

	// Health check endpoint
	r.Get(`/health`, healthHandler.HealthCheck)

	// Auth endpoints
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Get("/verify-email", authHandler.VerifyEmail)
		r.Post("/resend-verification", authHandler.ResendVerificationEmail)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh-token", authHandler.RefreshToken)
		r.Post("/logout", authHandler.Logout)
	})

	// Future user endpoints will be added here
	// Example:
	// r.Route("/api/users", func(r chi.Router) {
	//     r.Post("/", userHandler.CreateUser)
	//     r.Get("/{id}", userHandler.GetUser)
	// })

	return r
}
