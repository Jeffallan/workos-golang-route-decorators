package models

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type ResponseModel struct{
	Message string `json:"message"`
	ExtraContext string `json:"extra_context"`
	ClaimsMessage jwt.MapClaims  `json:"claims_message"`
}

type UserClaims struct {
	Sub string 
  	Sid string 
  	Jti string 
  	OrgId string
  	Role string
  	Permissions []string
  	Exp int64
  	Iat  int64
}

func IsAdminViaContext(u UserClaims, ctx context.Context) bool {
	return u.Role == "admin"
}

func HasPermission(u UserClaims, permission string) bool {
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}
// use this with url params: i.e /{:user_id}
func IsUserById(u UserClaims, userId string) bool {
	return u.Sub == userId
}

// use this with url params: i.e /{:org_id}
func IsOrgMemberById(u UserClaims, orgId string) bool {
	return u.OrgId == orgId
}
