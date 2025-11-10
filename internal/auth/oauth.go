package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// NewOAuthConfig creates a new OAuth2 configuration for Google
func NewOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  redirectURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/cloud-platform",
		},
		Endpoint: google.Endpoint,
	}
}
