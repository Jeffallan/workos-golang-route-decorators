package constants

type contextKey string

const claimsContextKey = contextKey("jwtClaims")

func GetJWTContextKey() string {
	return string(claimsContextKey)
}