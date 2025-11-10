# supreme-broccoli

A web-based cloud terminal application that provides authenticated access to Google Cloud Shell through a browser interface.

## Features

- Google OAuth2 authentication with session management
- WebSocket-based terminal emulation using xterm.js
- Reverse proxy to Theia IDE for cloud-based code editing
- MongoDB database for user token storage and role-based access control
- Interactive learning environment with lab guides, notes, and discussion panels

## Prerequisites

- Go 1.24 or higher
- MongoDB 4.4 or higher
- Google Cloud Platform account with OAuth2 credentials
- Google Cloud SDK (gcloud CLI)

## MongoDB Setup

### Local MongoDB Installation

**macOS (using Homebrew):**
```bash
brew tap mongodb/brew
brew install mongodb-community
brew services start mongodb-community
```

**Linux (Ubuntu/Debian):**
```bash
wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list
sudo apt-get update
sudo apt-get install -y mongodb-org
sudo systemctl start mongod
```

**Using Docker:**
```bash
docker run -d -p 27017:27017 --name mongodb \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin123 \
  mongo:6.0
```

### MongoDB Atlas (Cloud)

1. Create a free account at [MongoDB Atlas](https://www.mongodb.com/cloud/atlas)
2. Create a new cluster
3. Configure network access (add your IP address)
4. Create a database user with read/write permissions
5. Get your connection string from the "Connect" button

### Create Application Database User

Connect to MongoDB and create a dedicated user for the application:

```javascript
// Connect to MongoDB
mongosh "mongodb://localhost:27017" -u admin -p admin123 --authenticationDatabase admin

// Switch to authdb database
use authdb

// Create application user with readWrite permissions
db.createUser({
  user: "supreme-broccoli-app",
  pwd: "secure-password-here",
  roles: [
    { role: "readWrite", db: "authdb" }
  ]
})
```

### Required MongoDB User Permissions

The application requires a MongoDB user with the following permissions:

- **Database**: `authdb` (or your chosen database name)
- **Role**: `readWrite`
- **Operations**:
  - `find` - Read user documents
  - `insert` - Create new user documents
  - `update` - Update existing user tokens
  - `remove` - Delete user documents (if needed)

**Minimum privilege principle**: The application user should NOT have admin privileges or access to other databases.

## Configuration

### Environment Variables

Copy the `.env.example` to `.env` and configure:

```bash
# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret

# MongoDB Connection String
# Choose one of the following formats based on your setup:

# Local MongoDB (no authentication)
DB_DSN=mongodb://localhost:27017/authdb

# Local MongoDB (with authentication)
DB_DSN=mongodb://username:password@localhost:27017/authdb?authSource=admin

# MongoDB Atlas (cloud)
DB_DSN=mongodb+srv://username:password@cluster.mongodb.net/authdb?retryWrites=true&w=majority

# MongoDB with replica set
DB_DSN=mongodb://username:password@host1:27017,host2:27017/authdb?replicaSet=rs0&authSource=admin

# Session Management (generate a random 32-byte base64 string)
SESSION_KEY=your-random-session-key

# Application URL
APP_BASE_URL=http://localhost:8080
```

### MongoDB Connection String Format

The MongoDB connection string follows this format:

```
mongodb://[username:password@]host[:port][/database][?options]
```

**Components:**
- `mongodb://` - Protocol (use `mongodb+srv://` for Atlas)
- `username:password@` - Authentication credentials (optional for local dev)
- `host` - MongoDB server hostname or IP
- `port` - MongoDB port (default: 27017, omit for Atlas)
- `database` - Database name (e.g., `authdb`)
- `options` - Query parameters for connection configuration

**Common Options:**
- `authSource=admin` - Authentication database (usually `admin`)
- `retryWrites=true` - Enable retryable writes
- `w=majority` - Write concern level
- `tls=true` - Enable TLS/SSL encryption
- `replicaSet=rs0` - Replica set name

## Project Structure

The application follows a modular structure for better maintainability:

```
supreme-broccoli/
├── cmd/server/          # Application entry point
├── internal/
│   ├── auth/           # OAuth and session management
│   ├── config/         # Configuration loading
│   ├── database/       # MongoDB operations
│   ├── handlers/       # HTTP request handlers
│   ├── middleware/     # HTTP middleware
│   └── models/         # Data models
├── index.html          # Terminal UI
├── login.html          # Login page
└── Makefile           # Build automation
```

See [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) for detailed documentation.

## Installation

### Install Dependencies

```bash
make deps
# or
go mod download
```

### Build the Application

```bash
make build
# or
go build -o bin/supreme-broccoli cmd/server/main.go
```

## Running the Application

### Start the Application

```bash
make run
# or
go run cmd/server/main.go
```

The application will:
1. Load configuration from environment variables
2. Connect to MongoDB using the `DB_DSN` from `.env`
3. Verify database connectivity
4. Start the HTTP server on port 8080

### Access the Application

Open your browser and navigate to:
```
http://localhost:8080
```

## Available Make Commands

```bash
make build         # Build the application
make run           # Run the application (auto-loads .env)
make start         # Run the built binary (auto-loads .env)
make clean         # Clean build artifacts
make test          # Run tests
make migrate-build # Build migration tool
make migrate-run   # Run migration
make deps          # Install dependencies
make fmt           # Format code
make help          # Show all commands
```

**Note:** The `make run` and `make start` commands automatically load environment variables from the `.env` file.

## Data Migration

If migrating from MySQL to MongoDB, use the migration utility:

```bash
# Build the migration tool
go build -o migrate migrate.go

# Run the migration
./migrate
```

The migration tool will:
1. Connect to both MySQL (old) and MongoDB (new)
2. Read all users from MySQL
3. Transform and insert them into MongoDB
4. Verify the migration was successful
5. Report any errors

**Note**: Update your `.env` file with both connection strings before running the migration.

## Database Schema

### Users Collection

The application stores user data in the `users` collection with the following structure:

```json
{
  "_id": "user@example.com",
  "access_token": "ya29.a0AfH6SMB...",
  "refresh_token": "1//0gHdP9...",
  "token_expiry": ISODate("2025-11-10T15:30:00Z"),
  "role": "user"
}
```

**Fields:**
- `_id` (string) - User email address (unique identifier)
- `access_token` (string) - OAuth2 access token
- `refresh_token` (string) - OAuth2 refresh token
- `token_expiry` (datetime) - Token expiration timestamp
- `role` (string) - User role: `"user"` or `"admin"`

### Indexes

The `_id` field is automatically indexed by MongoDB. No additional indexes are required for basic operation.

## Development

### Update Dependencies

```bash
go get -u ./...
go mod tidy
```

### Run Tests

```bash
go test ./...
```

## Troubleshooting

### MongoDB Connection Issues

**Error**: `Failed to connect to MongoDB`

**Solutions:**
- Verify MongoDB is running: `mongosh --eval "db.adminCommand('ping')"`
- Check connection string format in `.env`
- Verify network access (firewall, security groups)
- For Atlas: Ensure your IP is whitelisted

**Error**: `Authentication failed`

**Solutions:**
- Verify username and password in connection string
- Check `authSource` parameter (usually `admin`)
- Ensure user exists: `db.getUsers()` in mongosh
- Verify user has correct permissions on `authdb`

### Application Issues

**Error**: `User not found`

**Solutions:**
- Verify user document exists in MongoDB
- Check email format matches exactly
- Run migration if coming from MySQL

**Error**: `Token expired`

**Solutions:**
- Application automatically refreshes tokens
- Verify refresh_token is stored in database
- Check Google OAuth credentials are valid

## License

MIT