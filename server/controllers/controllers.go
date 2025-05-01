package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter" // Import httprouter
	"workos-golang-route-decorators/server/constants"
	"workos-golang-route-decorators/server/models"
)

// RootHandler updated to httprouter.Handle signature
func RootHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) { // Added _ httprouter.Params

	var extraContextValue string
	if val, ok := r.Context().Value("extra_context").(string); ok {
		extraContextValue = val
	}

	// Safely get claims from context
	claims, ok := r.Context().Value(constants.GetJWTContextKey()).(jwt.MapClaims)
	if !ok {
		log.Printf("RootHandler: Failed to get jwt.MapClaims from context.")
		// Handle case where claims are missing - maybe return error or default claims
		claims = jwt.MapClaims{"error": "claims missing"} // Example default
	}


	res := models.ResponseModel{
		Message:       "Hello, from the server!",
		ExtraContext:  extraContextValue,
		ClaimsMessage: claims, // Assign the retrieved claims (or default)
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
