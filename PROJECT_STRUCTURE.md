# Project Structure

This document describes the reorganized project structure for supreme-broccoli.

## Directory Layout

```
supreme-broccoli/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── auth/
│   │   ├── oauth.go             # OAuth2 configuration
│   │   └── session.go           # Session management
│   ├── config/
│   │   └── config.go            # Configuration loading
│   ├── database/
│   │   └── mongodb.go           # MongoDB operations
│   ├── handlers/
│   │   ├── admin_handlers.go   # Admin route handlers
│   │   ├── auth_handlers.go    # Authentication handlers
│   │   ├── proxy_handlers.go   # Theia proxy handlers
│   │   └── terminal_handlers.go # Terminal/WebSocket handlers
│   ├── middleware/
│   │   └── auth.go              # Authentication middleware
│   └── models/
│       └── user.go              # User data model
├── .env                         # Environment variables
├── .gitignore                   # Git ignore rules
├── go.mod                       # Go module definition
├── go.sum                       # Go dependencies
├── index.html                   # Terminal UI
├── login.html                   # Login page
├── Makefile                     # Build automation
├── migrate.go                   # MySQL to MongoDB migration tool
└── README.md                    # Project documentation
```

## Package Descriptions

### `cmd/server`
Contains the application entry point. Initializes all components and starts the HTTP server.

### `internal/auth`
Handles OAuth2 configuration and session management.
- `oauth.go`: Creates Google OAuth2 configuration
- `session.go`: Manages cookie-based sessions

### `internal/config`
Loads and validates application configuration from environment variables.

### `internal/database`
Manages MongoDB connection and operations.
- Connection management
- User CRUD operations
- Error handling

### `internal/handlers`
HTTP request handlers organized by functionality:
- `auth_handlers.go`: Login, OAuth callback
- `terminal_handlers.go`: Terminal page, WebSocket connections
- `proxy_handlers.go`: Theia IDE reverse proxy
- `admin_handlers.go`: Admin-only routes

### `internal/middleware`
HTTP middleware for cross-cutting concerns:
- `auth.go`: Authentication and authorization checks

### `internal/models`
Data models and structures:
- `user.go`: User model with OAuth tokens

## Building and Running

### Using Make (Recommended)

```bash
# Build the application
make build

# Run the application
make run

# Clean build artifacts
make clean

# Run tests
make test

# Build migration tool
make migrate-build

# Run migration
make migrate-run

# Install dependencies
make deps

# Format code
make fmt

# Show all commands
make help
```

### Using Go Commands

```bash
# Build
go build -o bin/supreme-broccoli cmd/server/main.go

# Run
go run cmd/server/main.go

# Test
go test -v ./...
```

## Configuration

Set the following environment variables in `.env`:

```bash
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
DB_DSN=mongodb+srv://user:pass@cluster.mongodb.net/authdb
SESSION_KEY=your-session-key
APP_BASE_URL=http://localhost:8080
SERVER_PORT=8080  # Optional, defaults to 8080
```

## Key Improvements

### 1. Modular Structure
- Code organized into logical packages
- Clear separation of concerns
- Easy to test and maintain

### 2. Removed Auto-Installation Scripts
- No automatic Theia installation on Cloud Shell
- Cleaner WebSocket connection handling
- Manual setup gives users more control

### 3. Better Configuration Management
- Centralized configuration loading
- Environment variable validation
- Default values for optional settings

### 4. Improved Error Handling
- Consistent error messages
- Proper context timeouts
- Graceful shutdown

### 5. Build Automation
- Makefile for common tasks
- Easy to build and deploy
- Consistent development workflow

## Migration from Old Structure

The old `main.go` has been split into:
- Configuration → `internal/config/config.go`
- Database → `internal/database/mongodb.go`
- Auth logic → `internal/auth/` and `internal/handlers/auth_handlers.go`
- Terminal → `internal/handlers/terminal_handlers.go`
- Middleware → `internal/middleware/auth.go`
- Models → `internal/models/user.go`

## Development Workflow

1. **Make changes** to code in `internal/` or `cmd/`
2. **Format code**: `make fmt`
3. **Run tests**: `make test`
4. **Build**: `make build`
5. **Run locally**: `make run`
6. **Deploy**: Copy `bin/supreme-broccoli` to server

## Testing

Run tests with:
```bash
make test
```

Or for specific packages:
```bash
go test -v ./internal/database
go test -v ./internal/handlers
```

## Deployment

1. Build the binary:
   ```bash
   make build
   ```

2. Copy binary and static files to server:
   ```bash
   scp bin/supreme-broccoli user@server:/app/
   scp index.html login.html user@server:/app/
   ```

3. Set environment variables on server

4. Run the application:
   ```bash
   ./supreme-broccoli
   ```

## Notes

- The `internal/` directory prevents external packages from importing these modules
- All handlers receive dependencies via struct fields (dependency injection)
- MongoDB connection is properly closed on application shutdown
- Session store is configured with secure defaults
