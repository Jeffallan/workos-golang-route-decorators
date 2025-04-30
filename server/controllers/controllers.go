package controllers

import (
	"net/http"
	"encoding/json"
	"log"
	"github.com/golang-jwt/jwt/v5"
	"workos-golang-route-decorators/server/models"
	"workos-golang-route-decorators/server/constants"

)

func RootHandler(w http.ResponseWriter, r *http.Request) {

	var extraContextValue string
	if val, ok := r.Context().Value("extra_context").(string); ok {
		extraContextValue = val
	}

	res := models.ResponseModel{
		Message:       "Hello, from the server!",
		ExtraContext:  extraContextValue,
		ClaimsMessage: r.Context().Value(constants.GetJWTContextKey()).(jwt.MapClaims), 
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
