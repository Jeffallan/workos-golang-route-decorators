// /home/dmitry/projects/pocs/workos-golang-route-decorators/server/middleware/middleware.go
package middleware

import (
	"fmt"
	"log" // Added for logging errors
	"os"
	"time"

	"github.com/MicahParks/keyfunc/v2" // Import keyfunc
	"github.com/golang-jwt/jwt/v5"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

// VerifyToken validates a JWT string using the WorkOS JWKS endpoint.
func VerifyToken(tokenString string) (*jwt.Token, error) {
	// --- Get WorkOS Client ID ---
	clientID := os.Getenv("VITE_WORKOS_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("VITE_WORKOS_CLIENT_ID environment variable not set")
	}

	// --- Get JWKS URL ---
	apiKey := os.Getenv("WORKOS_API_KEY")
	if apiKey == "" {
		log.Println("Warning: WORKOS_API_KEY environment variable not set. JWKS URL fetch might fail if auth is required.")
		// Initialization of the SDK client should ideally happen once in main.go
		// and be passed around if needed, rather than relying on global state or
		// checking env vars repeatedly within request handling.
	}

	// Fetch the JWKS URL using the correct function name
	jwksURL, err := usermanagement.GetJWKSURL(clientID) // Pass clientID directly
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS URL: %w", err)
	}
	log.Printf("Using JWKS URL: %s", jwksURL) // Log the URL being used

	// --- Create JWKS Key Provider ---
	options := keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshErrorHandler: func(err error) {
			log.Printf("JWKS refresh error: %s", err)
		},
	}
	// Use jwksURL.String() to pass the URL string to keyfunc.Get
	jwks, err := keyfunc.Get(jwksURL.String(), options)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS keyfunc: %w", err)
	}
	defer jwks.EndBackground()

	// --- Parse and Validate JWT ---
	token, err := jwt.Parse(tokenString, jwks.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("token parsing/validation failed: %w", err)
	}

	// --- Additional Claim Validation (Recommended) ---
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 1. Check Issuer (iss)
		issuerHost := os.Getenv("VITE_WORKOS_API_HOSTNAME")
		expectedIssuer := fmt.Sprintf("https://%s", issuerHost) //"https://api.workos.com" // Example - VERIFY THIS with WorkOS docs!
		issuer, err := claims.GetIssuer()
		if err != nil {
			// Error means claim is missing or wrong type
			return nil, fmt.Errorf("issuer (iss) claim error: %w", err)
		}
		if issuer != expectedIssuer {
			return nil, fmt.Errorf("invalid issuer (iss): expected '%s', got '%s'", expectedIssuer, issuer)
		}

		// no audience is present in this jwt
		// // 2. Check Audience (aud) - Can be string or []string
		// audience, err := claims.GetAudience()
		// if err != nil {
		// 	// Error means claim is missing or wrong type
		// 	return nil, fmt.Errorf("audience (aud) claim error: %w", err)
		// }
		// // Check if the expected clientID is present in the audience list
		// fmt.Printf("token: %+v\n", token)
		// fmt.Printf("Audience: %v\n", audience)
		// foundAudience := false
		// for _, aud := range audience {
		// 	if aud == clientID {
		// 		foundAudience = true
		// 		break
		// 	}
		// }
		// if !foundAudience {
		// 	return nil, fmt.Errorf("invalid audience (aud): expected '%s' to be present, got '%v'", clientID, audience)
		// }

		// 3. Check Expiration (exp) - Usually handled by jwt.Parse, but explicit check is fine
		expTime, err := claims.GetExpirationTime()
		if err != nil {
			return nil, fmt.Errorf("expiration time (exp) claim error: %w", err)
		}
		// Use jwt.NewNumericDate(time.Now()) for comparison
		if expTime == nil || !expTime.After(time.Now()) {
			return nil, fmt.Errorf("token has expired (exp: %v)", expTime)
		}


		log.Println("Token signature and standard claims validated successfully.")
		return token, nil
	}

	return nil, fmt.Errorf("invalid token or claims format after parsing")
}
