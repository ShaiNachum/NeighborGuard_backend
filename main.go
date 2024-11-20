package main

import (
	"log"
	"neighborguard/api"
	"neighborguard/pkg/middleware"
	"net/http"

	_ "neighborguard/docs" // Import generated docs

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title NeighborGuard API
// @version 1.0
// @description This is the NeighborGuard API documentation.
// @host localhost:8080
// @BasePath /
func main() {
	const PORT string = "8080"
	router := mux.NewRouter()

	// Apply CORS middleware to all routes
	router.Use(middleware.CorsHandler)

	router = api.SetupRoutes(router)

	router.PathPrefix("/swagger").Handler(httpSwagger.Handler())

	log.Printf("Server running on port %s", PORT)
	http.ListenAndServe(":"+PORT, router)
}
