package auth

import (
	"encoding/gob"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"

	"supreme-broccoli/internal/models"
)

func init() {
	// Register types for session storage
	gob.Register(oauth2.Token{})
	gob.Register(models.User{})
}

// NewSessionStore creates a new cookie-based session store
func NewSessionStore(key string) *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte(key))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		// Secure: true, // Enable in production with HTTPS
	}
	return store
}
