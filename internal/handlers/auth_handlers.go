package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"supreme-broccoli/internal/database"
	"supreme-broccoli/internal/models"
)

type AuthHandlers struct {
	OAuthConfig  *oauth2.Config
	SessionStore *sessions.CookieStore
	DB           *database.MongoDB
}

// HandleLogin serves the login page
func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "login.html")
}

// HandleGoogleLogin starts the OAuth flow
func (h *AuthHandlers) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.OAuthConfig.AuthCodeURL("state-string", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback processes the OAuth callback
func (h *AuthHandlers) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.FormValue("code")

	// Exchange code for token
	token, err := h.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get user info
	client := h.OAuthConfig.Client(ctx, token)
	oauth2Service, err := oauth2api.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Failed to create oauth2 service: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Save user to database
	user := models.User{
		Email:        userInfo.Email,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
		Role:         "user",
	}

	// Preserve existing role if user exists
	existingUser, err := h.DB.GetUser(user.Email)
	if err == nil {
		user.Role = existingUser.Role
	}

	if err := h.DB.SaveUser(user); err != nil {
		log.Printf("Failed to save user to DB: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Create session
	session, _ := h.SessionStore.Get(r, "auth-session")
	session.Values["email"] = user.Email
	session.Values["role"] = user.Role
	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
	}

	http.Redirect(w, r, "/terminal/", http.StatusSeeOther)
}
