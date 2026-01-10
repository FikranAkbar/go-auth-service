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
	)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(envs.Server.WriteTimeout))

	// Health check endpoint
	r.Get(`/health`, healthHandler.HealthCheck)

	// Future user endpoints will be added here
	// Example:
	// r.Route("/api/v1/users", func(r chi.Router) {
	//     r.Post("/", userHandler.CreateUser)
	//     r.Get("/{id}", userHandler.GetUser)
	// })

	return r
}
