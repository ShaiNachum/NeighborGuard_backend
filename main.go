package main

import (
	"log"
	"neighborguard/api"
	"neighborguard/pkg/database"
	"neighborguard/pkg/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "neighborguard/docs" // Import generated docs

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title NeighborGuard API
// @version 1.0
// @description This is the NeighborGuard API documentation.
// @BasePath /
func main() {
	const PORT string = "8080"
	
	// Connect to MongoDB
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	// Ensure MongoDB connection is closed when the application is terminated
	defer database.Disconnect()
	
	// Create router
	router := mux.NewRouter()

	// Apply CORS middleware to all routes
	router.Use(middleware.CorsHandler)

	// Setup API routes
	router = api.SetupRoutes(router)

	// Setup Swagger documentation
	router.PathPrefix("/swagger").Handler(httpSwagger.Handler())

	// Start HTTP server
	log.Printf("Server running on port %s", PORT)
	
	// Create a server
	server := &http.Server{
		Addr:    ":" + PORT,
		Handler: router,
	}
	
	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Server shutting down...")
}