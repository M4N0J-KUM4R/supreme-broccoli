package main

import (
	"context"
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil" // Proxy
	"net/url"            // Proxy
	"os"
	"os/exec"
	"strings" // Proxy
	"time"

	"github.com/creack/pty"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// --- CONFIGURATION (Load from Environment Variables) ---
// DO NOT hardcode these values. Set them in your shell:
// export GOOGLE_CLIENT_ID="your-client-id.apps.googleusercontent.com"
// export GOOGLE_CLIENT_SECRET="your-client-secret"
// export DB_DSN="admin:your-new-password@tcp(database-1.c78g46gakxa8.eu-north-1.rds.amazonaws.com:3306)/your-db-name"
// export SESSION_KEY="a-very-strong-random-key"
// export APP_BASE_URL="http://localhost:8080" // Or your public URL

var (
	googleOauthConfig *oauth2.Config
	sessionStore      *sessions.CookieStore
	db                *sql.DB
)

// This is the old upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

// User struct to store token and user info
// We will store this in the database
type User struct {
	Email        string
	AccessToken  string
	RefreshToken string
	TokenExpiry  time.Time
	Role         string
}

// We need to register this struct so the session cookie can store it
func init() {
	gob.Register(oauth2.Token{})
	gob.Register(User{})
}

// --- DATABASE ---
/*
CREATE TABLE IF NOT EXISTS users (
    email VARCHAR(255) NOT NULL PRIMARY KEY,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expiry TIMESTAMP NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user'
);
*/

func initDB() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable not set.")
	}

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established.")
}

// Saves or updates a user's tokens in the database
func saveUserToken(user User) error {
	query := `
	INSERT INTO users (email, access_token, refresh_token, token_expiry, role)
	VALUES (?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		access_token = VALUES(access_token),
		refresh_token = VALUES(refresh_token),
		token_expiry = VALUES(token_expiry),
		role = VALUES(role);
	`
	_, err := db.Exec(query, user.Email, user.AccessToken, user.RefreshToken, user.TokenExpiry, user.Role)
	return err
}

// Gets a user's details from the database
func getUser(email string) (User, error) {
	var user User
	query := `SELECT email, access_token, refresh_token, token_expiry, role FROM users WHERE email = ?`
	err := db.QueryRow(query, email).Scan(&user.Email, &user.AccessToken, &user.RefreshToken, &user.TokenExpiry, &user.Role)
	return user, err
}

// --- HANDLERS ---

// handleLogin serves the login page
func handleLogin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "login.html")
}

// handleTerminal serves the protected terminal page
func handleTerminal(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "auth-session")
	email, ok := session.Values["email"].(string)
	if !ok || email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, "index.html")
}

// handleGoogleLogin starts the OAuth flow
func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state-string", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// handleGoogleCallback is where Google redirects back to
func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.FormValue("code")

	// 1. Exchange the code for a token
	token, err := googleOauthConfig.Exchange(ctx, code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// 2. Get user info (like email)
	client := googleOauthConfig.Client(ctx, token)
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

	// 3. Save user and tokens to database
	user := User{
		Email:        userInfo.Email,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
		Role:         "user", // Set default role
	}

	existingUser, err := getUser(user.Email)
	if err == nil {
		user.Role = existingUser.Role // Preserve existing role
	}

	if err := saveUserToken(user); err != nil {
		log.Printf("Failed to save user to DB: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// 4. Create a session
	session, _ := sessionStore.Get(r, "auth-session")
	session.Values["email"] = user.Email
	session.Values["role"] = user.Role
	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
	}

	// 5. Redirect to the protected terminal page
	http.Redirect(w, r, "/terminal/", http.StatusSeeOther)
}

// --- Proxy Handler for Theia Editor ---
func handleEditorProxy(w http.ResponseWriter, r *http.Request) {
	// The gcloud ssh command forwards Theia's 8080 port to our local 9090
	targetURL, err := url.Parse("http://localhost:9090")
	if err != nil {
		log.Printf("Failed to parse target URL: %v", err)
		http.Error(w, "Error parsing proxy URL", http.StatusInternalServerError)
		return
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Custom Director to fix headers for WebSocket and paths
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = r.URL.Path
		req.Host = targetURL.Host

		// Handle WebSocket upgrade headers
		if r.Header.Get("Connection") == "Upgrade" && r.Header.Get("Upgrade") == "websocket" {
			log.Println("Proxying WebSocket upgrade request to Theia")
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Upgrade", "websocket")
		}
	}

	// Custom ModifyResponse to handle redirects
	proxy.ModifyResponse = func(res *http.Response) error {
		// Theia might redirect. We need to rewrite the Location header
		if loc, ok := res.Header["Location"]; ok && len(loc) > 0 {
			if strings.HasPrefix(loc[0], targetURL.String()) {
				// Rewrite "http://localhost:9090/path" to "/editor/path"
				res.Header.Set("Location", strings.Replace(loc[0], targetURL.String(), "/editor", 1))
			} else if strings.HasPrefix(loc[0], "/") {
				// Rewrite "/path" to "/editor/path"
				res.Header.Set("Location", "/editor"+loc[0])
			}
		}
		return nil
	}

	log.Printf("Proxying editor request for: %s", r.URL.Path)
	proxy.ServeHTTP(w, r)
}

// serveWs handles the WebSocket connection
func serveWs(w http.ResponseWriter, r *http.Request) {
	log.Println("New WebSocket connection...")

	// 1. Check for a valid session
	session, err := sessionStore.Get(r, "auth-session")
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

	// 2. Get the user's tokens from the database
	user, err := getUser(email)
	if err != nil {
		log.Printf("Failed to get user %s from DB: %v", email, err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// We need a context for the token refresh logic
	ctx := context.Background()

	// --- START: Handle Token Refresh ---
	if time.Now().After(user.TokenExpiry.Add(-1 * time.Minute)) { // Refresh if 1 min left
		log.Printf("User %s token is expired or expiring, attempting refresh...", user.Email)
		token := &oauth2.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
			Expiry:       user.TokenExpiry,
		}
		tokenSource := googleOauthConfig.TokenSource(ctx, token)
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
			if err := saveUserToken(user); err != nil {
				log.Printf("Failed to save refreshed token to DB for %s: %v", user.Email, err)
			}
		} else {
			log.Printf("Token for %s was still valid.", user.Email)
		}
	}
	// --- END: Handle Token Refresh ---

	// 3. Upgrade the HTTP connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// 4. --- Prepare the 'gcloud' command ---
	log.Println("Starting gcloud ssh with port forwarding...")
	portForwardFlag := "-L 9090:localhost:8080" // Forwards remote 8080 to local 9090

	cmd := exec.Command("gcloud", "cloud-shell", "ssh",
		"--authorize-session",
		"--quiet",
		"--ssh-flag="+portForwardFlag, // Just port forward
	)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "CLOUDSDK_AUTH_ACCESS_TOKEN="+user.AccessToken)
	// --- END MODIFICATION ---

	// 5. Start the command in a pseudo-terminal (PTY)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("Failed to start pty: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Failed to start remote shell."))
		return
	}
	defer ptmx.Close()
	log.Println("PTY started successfully.")

	// --- NEW: Write verification commands to the PTY ---
	go func() {
		// Wait a brief moment for the shell to initialize and print its MOTD
		time.Sleep(2 * time.Second)

		// \n at the end of each command simulates pressing 'Enter'
		installCmd := "if ! command -v theia > /dev/null; then echo '--- Theia not found, attempting to install via npm... ---' && npm install -g @theia/cli && echo '--- Theia installed. ---'; else echo '--- Theia is already installed. ---'; fi\n"
		
		// --- MODIFIED: Create project dir and cd into it ---
		startCmd := "echo '--- Creating Theia project directory... ---' && mkdir -p my-theia-project && cd my-theia-project && echo '--- Starting Theia IDE... ---' && theia start --port=8080 --without-browser &\n"
		// --- END MODIFICATION ---

		log.Println("Writing install/start commands to PTY...")

		// Send the install/check command
		if _, err := ptmx.Write([]byte(installCmd)); err != nil {
			log.Printf("Failed to write install cmd to pty: %v", err)
			return
		}

		// Wait a moment for the check to complete
		time.Sleep(1 * time.Second)

		// Send the start command
		if _, err := ptmx.Write([]byte(startCmd)); err != nil {
			log.Printf("Failed to write start cmd to pty: %v", err)
			return
		}
	}()
	// --- END NEW ---

	// 6. Bridge PTY and WebSocket
	// Goroutine 1: Read from PTY and write to WebSocket
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

	// Goroutine 2: Read from WebSocket and write to PTY
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		if msgType == websocket.TextMessage {
			log.Println("Got text message:", string(msg))
		} else if msgType == websocket.BinaryMessage {
			if _, err := ptmx.Write(msg); err != nil {
				log.Println("PTY write error:", err)
				break
			}
		}
	}
	log.Println("WebSocket connection closed.")
}

// --- Middleware ---

// authMiddleware checks if a user is authenticated
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, "auth-session")
		email, ok := session.Values["email"].(string)

		if !ok || email == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// User is authenticated, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// adminMiddleware checks if a user is authenticated AND is an admin
func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, "auth-session")
		email, okEmail := session.Values["email"].(string)
		role, okRole := session.Values["role"].(string)

		if !okEmail || email == "" {
			// Not authenticated
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if !okRole || role != "admin" {
			// Authenticated, but not an admin
			http.Error(w, "Forbidden: You do not have admin privileges.", http.StatusForbidden)
			return
		}

		// User is authenticated and is an admin, proceed
		next.ServeHTTP(w, r)
	})
}

// --- Admin Page Handler ---

// handleAdmin serves a simple admin-only page
func handleAdmin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the Admin-Only Page!")
}

// --- Main Function ---

func main() {
	// --- Load Config ---
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	sessionKey := os.Getenv("SESSION_KEY")
	appBaseURL := os.Getenv("APP_BASE_URL")

	if googleClientID == "" || googleClientSecret == "" || sessionKey == "" || appBaseURL == "" {
		log.Fatal("GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, SESSION_KEY, and APP_BASE_URL must be set.")
	}

	// --- Init Database ---
	initDB()

	// --- Init OAuth ---
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  appBaseURL + "/auth/google/callback",
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/cloud-platform", // Scope for gcloud
		},
		Endpoint: google.Endpoint,
	}

	// --- Init Sessions ---
	sessionStore = sessions.NewCookieStore([]byte(sessionKey))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		// Secure: true, // Uncomment this in production (when using HTTPS)
	}

	// --- Register Routes ---

	// Public routes (no middleware)
	http.HandleFunc("/", handleLogin)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/auth/google", handleGoogleLogin)
	http.HandleFunc("/auth/google/callback", handleGoogleCallback)

	// Protected routes for all authenticated users
	// This now acts as a prefix route for our single-page app
	http.Handle("/terminal/", authMiddleware(http.HandlerFunc(handleTerminal)))

	// The WebSocket route's auth is handled *inside* serveWs
	http.HandleFunc("/ws", serveWs)

	// Protected routes for admin users only
	http.Handle("/admin", adminMiddleware(http.HandlerFunc(handleAdmin)))

	// --- Add Proxy Route ---
	// The proxy must also be behind the auth middleware
	// We strip /editor/ so the proxy passes the correct path to Theia
	proxyHandler := http.StripPrefix("/editor/", http.HandlerFunc(handleEditorProxy))
	http.Handle("/editor/", authMiddleware(proxyHandler))

	log.Println("Starting web terminal server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}