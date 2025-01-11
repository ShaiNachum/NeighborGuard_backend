package api

import (
	"neighborguard/api/handlers"
	"neighborguard/pkg/middleware"

	"github.com/gorilla/mux"
)

// SetupRoutes sets up the routes for the API
func SetupRoutes(router *mux.Router) *mux.Router {
	// Health check endpoint
	router.HandleFunc("/healthz", middleware.Chain(handlers.HealthHandler, middleware.Logging())).Methods("GET")

	// Collection endpoints (plural)
	router.HandleFunc("/users", middleware.Chain(handlers.GetUsers, middleware.Logging())).Methods("GET")
	router.HandleFunc("/users/recipients", middleware.Chain(handlers.GetNearbyRecipients, middleware.Logging())).Methods("GET")

	// Single user endpoints (singular)
	router.HandleFunc("/user", middleware.Chain(handlers.CreateUser, middleware.Logging())).Methods("POST")
	router.HandleFunc("/user/{email}", middleware.Chain(handlers.GetUserByEmail, middleware.Logging())).Methods("GET")
	router.HandleFunc("/user/{uid}", middleware.Chain(handlers.UpdateUser, middleware.Logging())).Methods("PUT")

	// Meeting endpoints
	router.HandleFunc("/meeting", middleware.Chain(handlers.CreateMeeting, middleware.Logging())).Methods("POST")
	router.HandleFunc("/meeting/{id}", middleware.Chain(handlers.CancelMeeting, middleware.Logging())).Methods("DELETE")
	router.HandleFunc("/meeting/{id}/status", middleware.Chain(handlers.UpdateMeetingStatus, middleware.Logging())).Methods("PUT")

	// Collection endpoint for getting meetings
	router.HandleFunc("/meetings", middleware.Chain(handlers.GetMeetings, middleware.Logging())).Methods("GET")

	return router
}
