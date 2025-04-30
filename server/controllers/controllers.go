package controllers

import (
	"net/http"
	"workos-golang-route-decorators/server/models"
)

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
