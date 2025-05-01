package wrappers

import (
	"context"
	"fmt"
	"net/http"
	"workos-golang-route-decorators/server/constants"
	"workos-golang-route-decorators/server/models"
	"github.com/julienschmidt/httprouter"
	"github.com/golang-jwt/jwt/v5"
)


func jwtMapToUserClaims(ctx context.Context) models.UserClaims{
	user := models.UserClaims{
		Claims: ctx.Value(constants.GetJWTContextKey()).(jwt.MapClaims),
	}
	return user
}

func BasicWrapper(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("Lets add some extra context to our decorator")

		ctx := context.WithValue(r.Context(), "extra_context", "context added via wrapper")

		addedCtx := r.WithContext(ctx)
		
		if false  {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		f(w,addedCtx)
	}
}


func IsAdmin(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := jwtMapToUserClaims(r.Context())

		if user.IsAdmin() {
			f(w, r)
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}	
}

func IsUserViaURL(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := jwtMapToUserClaims(r.Context())
		params := httprouter.ParamsFromContext(r.Context())
		userId := params.ByName("user_id")
		fmt.Printf("\n\nUSER: %v\n%+v\n%+v\n", userId, r.Context(), params)
		if user.IsUserById(userId) {
			f(w, r)
			return
		}
		
		http.Error(w, "Unauthorized", http.StatusUnauthorized)	
	}	
}
func IsOrgMemberViaURL(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := jwtMapToUserClaims(r.Context())
		params := httprouter.ParamsFromContext(r.Context())
		orgId := params.ByName("org_id")
		fmt.Printf("\n\nORG: %v\n\n", orgId)
		if user.IsOrgMemberById(orgId) {
			f(w, r)
			return
		}
		
		http.Error(w, "Unauthorized", http.StatusUnauthorized)	
	}	
}