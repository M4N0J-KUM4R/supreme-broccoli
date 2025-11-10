package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Auth checks if a user is authenticated
func Auth(sessionStore *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := sessionStore.Get(r, "auth-session")
			email, ok := session.Values["email"].(string)

			if !ok || email == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Admin checks if a user is authenticated AND is an admin
func Admin(sessionStore *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := sessionStore.Get(r, "auth-session")
			email, okEmail := session.Values["email"].(string)
			role, okRole := session.Values["role"].(string)

			if !okEmail || email == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			if !okRole || role != "admin" {
				http.Error(w, "Forbidden: You do not have admin privileges.", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
