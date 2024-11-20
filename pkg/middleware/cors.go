package middleware

import (
	"net/http"
)

// func CorsHandler() Middleware {

// 	return func(f http.HandlerFunc) http.HandlerFunc {
// 		c := cors.New(cors.Options{
// 			AllowedOrigins:   []string{"*"}, // Allow all origins
// 			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
// 			AllowedHeaders:   []string{"*"}, // Allow all headers
// 			AllowCredentials: true,
// 		})
// 		// Call the next middleware/handler in chain
// 		return http.HandlerFunc(c.Handler(f).ServeHTTP)
// 	}
// }

// CorsHandler sets the CORS headers
func CorsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
