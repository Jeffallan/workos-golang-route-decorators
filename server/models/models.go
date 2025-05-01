package models

import (
	"github.com/golang-jwt/jwt/v5"
)

type ResponseModel struct{
	Message string `json:"message"`
	ExtraContext string `json:"extra_context"`
	ClaimsMessage jwt.MapClaims  `json:"claims_message"`
}

type UserClaims struct {
	Claims jwt.MapClaims
}

func (u UserClaims) IsAdmin() bool {
	return u.Claims["role"] == "admin"
}

func (u UserClaims) HasPermission(permission string) bool {
	for _, p := range u.Claims["permissions"].([]interface{}) {
		if p == permission {
			return true
		}
	}
	return false
}
// use this with url params: i.e /{:user_id}
func (u UserClaims) IsUserById(userId string) bool {
	return u.Claims["sub"] == userId
}

// use this with url params: i.e /{:org_id}
func (u UserClaims) IsOrgMemberById(orgId string) bool {
	return u.Claims["org_id"] == orgId
}
