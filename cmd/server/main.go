package main

import (
	"log"
	"net/http"

	"supreme-broccoli/internal/auth"
	"supreme-broccoli/internal/config"
	"supreme-broccoli/internal/database"
	"supreme-broccoli/internal/handlers"
	"supreme-broccoli/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.MongoDBURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize OAuth configuration
	oauthConfig := auth.NewOAuthConfig(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.AppBaseURL+"/auth/google/callback",
	)

	// Initialize session store
	sessionStore := auth.NewSessionStore(cfg.SessionKey)

	// Initialize handlers
	authHandlers := &handlers.AuthHandlers{
		OAuthConfig:  oauthConfig,
		SessionStore: sessionStore,
		DB:           db,
	}

	terminalHandlers := &handlers.TerminalHandlers{
		OAuthConfig:  oauthConfig,
		SessionStore: sessionStore,
		DB:           db,
	}

	// Initialize middleware
	authMiddleware := middleware.Auth(sessionStore)
	adminMiddleware := middleware.Admin(sessionStore)

	// Register routes
	// Public routes
	http.HandleFunc("/", authHandlers.HandleLogin)
	http.HandleFunc("/login", authHandlers.HandleLogin)
	http.HandleFunc("/auth/google", authHandlers.HandleGoogleLogin)
	http.HandleFunc("/auth/google/callback", authHandlers.HandleGoogleCallback)

	// Protected routes
	http.Handle("/terminal/", authMiddleware(http.HandlerFunc(terminalHandlers.HandleTerminal)))
	http.HandleFunc("/ws", terminalHandlers.HandleWebSocket)

	// Admin routes
	http.Handle("/admin", adminMiddleware(http.HandlerFunc(handlers.HandleAdmin)))

	// Editor proxy route
	proxyHandler := http.StripPrefix("/editor/", http.HandlerFunc(handlers.HandleEditorProxy))
	http.Handle("/editor/", authMiddleware(proxyHandler))

	// Start server
	log.Printf("Starting web terminal server on http://localhost:%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
