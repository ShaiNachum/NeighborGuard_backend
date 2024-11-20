package api

import (
	"neighborguard/api/handlers"
	"neighborguard/pkg/middleware"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router) *mux.Router {
	router.HandleFunc("/healthz", middleware.Chain(handlers.HealthHandler, middleware.Logging())).Methods("GET")

	router.HandleFunc("/users", middleware.Chain(handlers.GetUsers, middleware.Logging())).Methods("GET")
	router.HandleFunc("/users", middleware.Chain(handlers.CreateUser, middleware.Logging())).Methods("POST")

	return router
}
