package constants

type contextKey string

const claimsContextKey = contextKey("jwtClaims")

func GetJWTContextKey() string {
	return string(claimsContextKey)
}

func GetOrgID() string {
	return "org_id"
}

func GetUserID() string {
	return "user_id"
}