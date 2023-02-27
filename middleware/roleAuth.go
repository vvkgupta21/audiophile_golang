package middleware

import (
	"audio_phile/model"
	"fmt"
	"net/http"
)

type ContextKeys string

const (
	UserContext ContextKeys = "userInfo"
)

func UserContextData(r *http.Request) (string, error) {
	user := r.Context().Value(UserContext).(map[string]interface{})
	fmt.Println(user)
	var role string
	role = user["role"].(string)
	fmt.Println(role)
	return role, nil
}

func AdminMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, err := UserContextData(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println(role)
		if model.Role(role) == model.RoleAdmin {
			handler.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusForbidden)
	})
}

func UserMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, err := UserContextData(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println(role)
		if model.Role(role) == model.RoleUser {
			handler.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusForbidden)
	})
}
