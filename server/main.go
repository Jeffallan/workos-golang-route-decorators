package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"workos-golang-route-decorators/server/middleware"
	"workos-golang-route-decorators/server/wrappers"
	"workos-golang-route-decorators/server/constants"
	"workos-golang-route-decorators/server/controllers"
)

// --- Logging Middleware (Existing) ---
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("--> %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("<-- %s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// --- JWT Authentication Middleware (Existing) ---
func jwtAuthMiddleware(next http.Handler) http.Handler {
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
			ctx := context.WithValue(r.Context(), constants.GetJWTContextKey(), claims)
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


func main() {

	envErr := godotenv.Load("../.env.local")
	if envErr != nil {
		log.Printf("Warning: Could not load .env.local file: %v", envErr)
	}

	// --- Handler Route Registration ---
	router := httprouter.New()

	// Register routes directly using the refactored httprouter.Handle functions
	// The nesting order defines the execution flow (outermost wrapper first)
	router.GET("/", wrappers.BasicWrapper(controllers.RootHandler))
	router.GET("/admin", wrappers.BasicWrapper(wrappers.IsAdmin(controllers.RootHandler))) // Apply BasicWrapper then IsAdmin
	router.GET("/user/:user_id", wrappers.IsUserViaURL(controllers.RootHandler)) // Apply IsUserViaURL directly
	router.GET("/org/:org_id", wrappers.IsOrgMemberViaURL(controllers.RootHandler)) // Apply IsOrgMemberViaURL directly

	// --- CORS Configuration ---
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            false,
	})

	// --- Global Middleware Setup (Order Matters!) ---
	var handler http.Handler = router
	handler = jwtAuthMiddleware(handler) // Applied first to add claims to context
	handler = loggingMiddleware(handler)
	handler = c.Handler(handler)         // CORS applied outermost

	port := "8080"
	log.Printf("Starting Go server with CORS, JWT Auth, and httprouter enabled on http://localhost:%s\n", port)

	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
