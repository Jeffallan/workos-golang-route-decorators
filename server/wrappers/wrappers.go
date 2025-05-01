package wrappers

import (
	"context"
	"encoding/json" // Import encoding/json
	"fmt"
	"log" // Added for logging
	"net/http"
	"workos-golang-route-decorators/server/constants"
	"workos-golang-route-decorators/server/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter" // Import httprouter
)

// Helper function to send JSON error responses
func sendJSONErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	// Attempt to get extra context, default to empty string if not found
	var extraContextValue string
	if val, ok := r.Context().Value("extra_context").(string); ok {
		extraContextValue = val
	}

	// Create the error response struct
	errorResponse := models.ResponseModel{
		Message:       message,
		ExtraContext:  extraContextValue,
		ClaimsMessage: nil, // Set claims to nil for error responses
	}

	// Marshal the response
	jsonData, err := json.Marshal(errorResponse)
	if err != nil {
		// If marshaling fails, log the error and send a generic server error
		log.Printf("Error marshaling JSON error response: %v", err)
		http.Error(w, `{"message":"Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	// Set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode) // Set the desired status code (e.g., 403 Forbidden)
	_, err = w.Write(jsonData)
	if err != nil {
		// Log error if writing the response fails
		log.Printf("Error writing JSON error response: %v", err)
	}
}

// jwtMapToUserClaims remains the same
func jwtMapToUserClaims(ctx context.Context) models.UserClaims {
	claims, ok := ctx.Value(constants.GetJWTContextKey()).(jwt.MapClaims)
	if !ok {
		log.Printf("jwtMapToUserClaims: Failed to get jwt.MapClaims from context or incorrect type.")
		return models.UserClaims{Claims: jwt.MapClaims{}}
	}
	user := models.UserClaims{
		Claims: claims,
	}
	return user
}

// BasicWrapper updated to httprouter.Handle signature
func BasicWrapper(f httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), "extra_context", "context added via wrapper")
		addedCtx := r.WithContext(ctx)
		f(w, addedCtx, ps)
	}
}

// IsAdmin updated to httprouter.Handle signature
func IsAdmin(f httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user := jwtMapToUserClaims(r.Context())

		if len(user.Claims) == 0 {
			log.Println("IsAdmin Wrapper: Claims map is empty, cannot determine role.")
			// Use helper for 401 Unauthorized when claims are missing/invalid
			sendJSONErrorResponse(w, r, http.StatusUnauthorized, "Unauthorized: Invalid user claims")
			return
		}

		if user.IsAdmin() {
			f(w, r, ps)
			return
		}

		// Log failure details
		sub, _ := user.Claims["sub"].(string)
		role, _ := user.Claims["role"]
		log.Printf("IsAdmin Wrapper: User '%s' denied access (not admin). Role: '%v'", sub, role)
		// Use helper for 403 Forbidden
		sendJSONErrorResponse(w, r, http.StatusForbidden, "Forbidden: Admin privileges required")
	}
}

// IsUserViaURL updated to httprouter.Handle signature and uses ps directly
func IsUserViaURL(f httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user := jwtMapToUserClaims(r.Context())
		userId := ps.ByName("user_id")

		if len(user.Claims) == 0 {
			log.Println("IsUserViaURL Wrapper: Claims map is empty, cannot verify user ID.")
			// Use helper for 401 Unauthorized when claims are missing/invalid
			sendJSONErrorResponse(w, r, http.StatusUnauthorized, "Unauthorized: Invalid user claims")
			return
		}

		sub, _ := user.Claims["sub"].(string)
		fmt.Printf("\n\nIsUserViaURL: Checking if user '%s' matches URL param '%s'\n\n", sub, userId)

		if user.IsUserById(userId) {
			log.Printf("IsUserViaURL: User '%s' authorized for user_id '%s'.", sub, userId)
			f(w, r, ps)
			return
		}

		log.Printf("IsUserViaURL: User '%s' denied access for user_id '%s'.", sub, userId)
		// Use helper for 403 Forbidden
		sendJSONErrorResponse(w, r, http.StatusForbidden, "Forbidden: User mismatch")
	}
}

// IsOrgMemberViaURL updated to httprouter.Handle signature and uses ps directly
func IsOrgMemberViaURL(f httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user := jwtMapToUserClaims(r.Context())
		orgId := ps.ByName("org_id")

		if len(user.Claims) == 0 {
			log.Println("IsOrgMemberViaURL Wrapper: Claims map is empty, cannot verify org ID.")
			// Use helper for 401 Unauthorized when claims are missing/invalid
			sendJSONErrorResponse(w, r, http.StatusUnauthorized, "Unauthorized: Invalid user claims")
			return
		}

		sub, _ := user.Claims["sub"].(string)
		claimOrgId, _ := user.Claims["org_id"].(string)
		fmt.Printf("\n\nIsOrgMemberViaURL: Checking if user '%s' (org '%s') matches URL param '%s'\n\n", sub, claimOrgId, orgId)

		if user.IsOrgMemberById(orgId) {
			log.Printf("IsOrgMemberViaURL: User '%s' authorized for org_id '%s'.", sub, orgId)
			f(w, r, ps)
			return
		}

		log.Printf("IsOrgMemberViaURL: User '%s' (org '%s') denied access for org_id '%s'.", sub, claimOrgId, orgId)
		// Use helper for 403 Forbidden
		sendJSONErrorResponse(w, r, http.StatusForbidden, "Forbidden: Organization mismatch")
	}
}