package wrappers

import (
	"net/http"
	"fmt"
	"context"
)


func BasicWrapper(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Lets add some extra context to our decorator")

		ctx := context.WithValue(r.Context(), "extra_context", "context added via wrapper")

		addedCtx := r.WithContext(ctx)
		
		if false  {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		f(w,addedCtx)
	}
}