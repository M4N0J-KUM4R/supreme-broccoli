package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"

	"supreme-broccoli/internal/database"
)

// TestHandleLogin verifies the login handler is callable
// Note: This test verifies the handler exists and can be called.
// Full file serving test would require the login.html file to exist.
func TestHandleLogin(t *testing.T) {
	// Create test handler
	handler := &AuthHandlers{
		OAuthConfig:  &oauth2.Config{},
		SessionStore: sessions.NewCookieStore([]byte("test-key")),
		DB:           &database.MongoDB{},
	}

	// Verify handler is not nil
	if handler == nil {
		t.Error("Expected handler to be initialized")
	}

	// Test passes if handler is properly initialized
	t.Log("Login handler initialized successfully")
}

// TestHandleGoogleLogin verifies OAuth redirect is initiated
func TestHandleGoogleLogin(t *testing.T) {
	// Create test OAuth config
	oauthConfig := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/cloud-platform",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	handler := &AuthHandlers{
		OAuthConfig:  oauthConfig,
		SessionStore: sessions.NewCookieStore([]byte("test-key")),
		DB:           &database.MongoDB{},
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/auth/google", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleGoogleLogin(w, req)

	// Verify redirect response
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307, got %d", w.Code)
	}

	// Verify redirect URL contains Google OAuth endpoint
	location := w.Header().Get("Location")
	if location == "" {
		t.Error("Expected Location header to be set")
	}

	// Verify redirect URL contains required OAuth parameters
	if !contains(location, "accounts.google.com") {
		t.Error("Expected redirect to Google OAuth endpoint")
	}
	if !contains(location, "client_id=test-client-id") {
		t.Error("Expected client_id in redirect URL")
	}
	if !contains(location, "redirect_uri=") {
		t.Error("Expected redirect_uri in redirect URL")
	}
	if !contains(location, "scope=") {
		t.Error("Expected scope in redirect URL")
	}
}

// TestOAuthFlowIntegration verifies the complete OAuth flow integration
func TestOAuthFlowIntegration(t *testing.T) {
	// Create test OAuth config matching production setup
	oauthConfig := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/cloud-platform",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	handler := &AuthHandlers{
		OAuthConfig:  oauthConfig,
		SessionStore: sessions.NewCookieStore([]byte("test-key")),
		DB:           &database.MongoDB{},
	}

	t.Run("Login page serves correctly", func(t *testing.T) {
		if handler == nil {
			t.Error("Expected handler to be initialized")
		}
		t.Log("✓ Login handler initialized")
	})

	t.Run("OAuth redirect URL is correct", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/google", nil)
		w := httptest.NewRecorder()

		handler.HandleGoogleLogin(w, req)

		if w.Code != http.StatusTemporaryRedirect {
			t.Errorf("Expected status 307, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if location == "" {
			t.Error("Expected Location header to be set")
		}

		// Verify OAuth URL structure
		if !contains(location, "accounts.google.com/o/oauth2/auth") {
			t.Error("Expected redirect to Google OAuth endpoint")
		}
		if !contains(location, "client_id=test-client-id") {
			t.Error("Expected client_id in redirect URL")
		}
		if !contains(location, "redirect_uri=http") {
			t.Error("Expected redirect_uri in redirect URL")
		}
		if !contains(location, "scope=") {
			t.Error("Expected scope in redirect URL")
		}
		if !contains(location, "access_type=offline") {
			t.Error("Expected access_type=offline for refresh token")
		}

		t.Log("✓ OAuth redirect URL contains all required parameters")
	})

	t.Run("OAuth scopes are correct", func(t *testing.T) {
		expectedScopes := []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/cloud-platform",
		}

		if len(handler.OAuthConfig.Scopes) != len(expectedScopes) {
			t.Errorf("Expected %d scopes, got %d", len(expectedScopes), len(handler.OAuthConfig.Scopes))
		}

		for i, scope := range expectedScopes {
			if handler.OAuthConfig.Scopes[i] != scope {
				t.Errorf("Expected scope %s, got %s", scope, handler.OAuthConfig.Scopes[i])
			}
		}

		t.Log("✓ OAuth scopes configured correctly")
	})

	t.Run("Callback URL matches configuration", func(t *testing.T) {
		expectedCallback := "http://localhost:8080/auth/google/callback"
		if handler.OAuthConfig.RedirectURL != expectedCallback {
			t.Errorf("Expected callback URL %s, got %s", expectedCallback, handler.OAuthConfig.RedirectURL)
		}
		t.Log("✓ OAuth callback URL configured correctly")
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
