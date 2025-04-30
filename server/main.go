// server/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"workos-golang-route-decorators/server/middleware" 
	"workos-golang-route-decorators/server/models"
	"workos-golang-route-decorators/server/wrappers"
)

// --- Context Key for JWT Claims ---
type contextKey string

const claimsContextKey = contextKey("jwtClaims")

// --- Logging Middleware (Existing) ---
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("--> %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("<-- %s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// --- JWT Authentication Middleware ---
func jwtAuthMiddleware() func(http.Handler) http.Handler {
	// Return the actual middleware handler function
	return func(next http.Handler) http.Handler {
		// Return the http.HandlerFunc that processes each request
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("JWT Auth Middleware: Checking token...")

			// 1. Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Println("JWT Auth Middleware: Missing Authorization header")
				http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
				return // Stop processing
			}

			// 2. Check format "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				log.Println("JWT Auth Middleware: Invalid Authorization header format")
				http.Error(w, "Unauthorized: Invalid Authorization header format", http.StatusUnauthorized)
				return // Stop processing
			}
			tokenString := parts[1]

			// 3. Verify the token using the function from middleware package
			token, err := middleware.VerifyToken(tokenString)
			if err != nil {
				log.Printf("JWT Auth Middleware: Token validation failed: %v", err)
				http.Error(w, fmt.Sprintf("Unauthorized: %s", err.Error()), http.StatusUnauthorized)
				return // Stop processing
			}

			// 4. Token is valid, extract claims
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				log.Printf("JWT Auth Middleware: Token validated successfully. Claims: %v", claims)
				// Add claims to request context
				ctx := context.WithValue(r.Context(), claimsContextKey, claims)
				// Create a new request with the updated context
				newReq := r.WithContext(ctx)
				// Call the next handler with the new request
				next.ServeHTTP(w, newReq)
			} else {
				log.Println("JWT Auth Middleware: Invalid token claims or token marked invalid after parsing")
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return // Stop processing
			}
		})
	}
}


func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var extraContextValue string
	if val, ok := r.Context().Value("extra_context").(string); ok {
		extraContextValue = val
	}

	res := models.ResponseModel{
		Message:       "Hello, from the server!",
		ExtraContext:  extraContextValue,
		ClaimsMessage: r.Context().Value(claimsContextKey).(jwt.MapClaims), 
	}

	jsonData, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func main() {

	envErr := godotenv.Load("../.env.local")
	if envErr != nil {
		log.Printf("Warning: Could not load .env.local file: %v", envErr)
	}

	// --- Handler Route Registration ---
	http.HandleFunc("/", wrappers.BasicWrapper(rootHandler))

	// --- CORS Configuration ---
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true,
	})

	// --- Global Middleware Setup (Order Matters!) ---
	var handler http.Handler = http.DefaultServeMux
	handler = jwtAuthMiddleware()(handler)
	handler = loggingMiddleware(handler)
	handler = c.Handler(handler)

	port := "8080"
	log.Printf("Starting Go server with CORS and JWT Auth enabled on http://localhost:%s\n", port)

	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
