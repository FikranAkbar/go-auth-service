package router

import (
	"go-auth-service/internal/config"
	"go-auth-service/internal/handler"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitRouter(envs *config.Env, handlers ...interface{}) *chi.Mux {
	var (
		_ = handlers[0].(*handler.UserHandler)
	)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(envs.Server.WriteTimeout))

	r.Get(`/health`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"up"}`))
	})

	return r
}
