package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"

	"supreme-broccoli/internal/database"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type TerminalHandlers struct {
	OAuthConfig  *oauth2.Config
	SessionStore *sessions.CookieStore
	DB           *database.MongoDB
}

// HandleTerminal serves the terminal page
func (h *TerminalHandlers) HandleTerminal(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "auth-session")
	email, ok := session.Values["email"].(string)
	if !ok || email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, "index.html")
}

// HandleWebSocket manages WebSocket connections for the terminal
func (h *TerminalHandlers) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println("New WebSocket connection...")

	// Check for valid session
	session, err := h.SessionStore.Get(r, "auth-session")
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	email, okEmail := session.Values["email"].(string)
	if !okEmail || email == "" {
		log.Println("WebSocket connection without valid session.")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's tokens from database
	user, err := h.DB.GetUser(email)
	if err != nil {
		log.Printf("Failed to get user %s from DB: %v", email, err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	// Handle token refresh if needed
	if time.Now().After(user.TokenExpiry.Add(-1 * time.Minute)) {
		log.Printf("User %s token is expired or expiring, attempting refresh...", user.Email)
		token := &oauth2.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
			Expiry:       user.TokenExpiry,
		}
		tokenSource := h.OAuthConfig.TokenSource(ctx, token)
		newToken, err := tokenSource.Token()
		if err != nil {
			log.Printf("Failed to refresh token for %s: %v", user.Email, err)
			http.Error(w, "Failed to refresh session token", http.StatusUnauthorized)
			return
		}
		if newToken.AccessToken != user.AccessToken {
			log.Printf("Token successfully refreshed for %s", user.Email)
			user.AccessToken = newToken.AccessToken
			if newToken.RefreshToken != "" {
				user.RefreshToken = newToken.RefreshToken
			}
			user.TokenExpiry = newToken.Expiry
			if err := h.DB.SaveUser(user); err != nil {
				log.Printf("Failed to save refreshed token to DB for %s: %v", user.Email, err)
			}
		}
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Start gcloud command with port forwarding
	log.Println("Starting gcloud ssh with port forwarding...")
	portForwardFlag := "-L 9090:localhost:8080"

	cmd := exec.Command("gcloud", "cloud-shell", "ssh",
		"--authorize-session",
		"--quiet",
		"--ssh-flag="+portForwardFlag,
	)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "CLOUDSDK_AUTH_ACCESS_TOKEN="+user.AccessToken)

	// Start command in PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("Failed to start pty: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Failed to start remote shell."))
		return
	}
	defer ptmx.Close()
	log.Println("PTY started successfully.")

	// Bridge PTY and WebSocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				log.Println("PTY read error:", err)
				conn.Close()
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Println("WebSocket write error:", err)
				return
			}
		}
	}()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		if msgType == websocket.BinaryMessage {
			if _, err := ptmx.Write(msg); err != nil {
				log.Println("PTY write error:", err)
				break
			}
		}
	}
	log.Println("WebSocket connection closed.")
}
