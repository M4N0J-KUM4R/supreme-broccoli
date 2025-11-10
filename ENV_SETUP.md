# Environment Setup Guide

## Overview

The application requires several environment variables to run. These are automatically loaded from the `.env` file when using `make run` or `make start`.

## Required Environment Variables

### 1. Google OAuth Credentials

```bash
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
```

**How to get these:**
1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project or select existing one
3. Enable Google+ API
4. Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client ID"
5. Application type: "Web application"
6. Authorized redirect URIs: `http://localhost:8080/auth/google/callback`
7. Copy the Client ID and Client Secret

### 2. MongoDB Connection String

```bash
DB_DSN=mongodb+srv://user:pass@cluster.mongodb.net/authdb?retryWrites=true&w=majority
```

**Options:**

**Local MongoDB (no auth):**
```bash
DB_DSN=mongodb://localhost:27017/authdb
```

**Local MongoDB (with auth):**
```bash
DB_DSN=mongodb://username:password@localhost:27017/authdb?authSource=admin
```

**MongoDB Atlas (cloud):**
```bash
DB_DSN=mongodb+srv://username:password@cluster.mongodb.net/authdb?retryWrites=true&w=majority
```

### 3. Session Key

```bash
SESSION_KEY=your-random-32-byte-base64-string
```

**Generate a secure session key:**
```bash
# macOS/Linux
openssl rand -base64 32

# Or use any random 32+ character string
```

### 4. Application Base URL

```bash
APP_BASE_URL=http://localhost:8080
```

**For production:**
```bash
APP_BASE_URL=https://yourdomain.com
```

### 5. Server Port (Optional)

```bash
SERVER_PORT=8080
```

Default is 8080 if not specified.

## Complete .env Example

```bash
# Google OAuth Configuration
GOOGLE_CLIENT_ID=123456789-abcdefghijklmnop.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-AbCdEfGhIjKlMnOpQrStUvWx

# MongoDB Connection
DB_DSN=mongodb+srv://myuser:mypassword@cluster0.mongodb.net/authdb?retryWrites=true&w=majority

# Session Management
SESSION_KEY=VQx3JpAo7A30Ptm5NWVXTXtw9oK9RgPNvLzx55zgNKQ=

# Application Configuration
APP_BASE_URL=http://localhost:8080
SERVER_PORT=8080
```

## How Environment Variables Are Loaded

### Using Make Commands (Recommended)

The Makefile automatically loads `.env`:

```bash
make run    # Loads .env and runs with go run
make start  # Loads .env and runs the binary
```

**How it works:**
```makefile
ifneq (,$(wildcard ./.env))
    include .env
    export
endif
```

### Manual Loading

If running without Make:

**Option 1: Export variables**
```bash
export $(cat .env | grep -v '^#' | xargs)
go run cmd/server/main.go
```

**Option 2: Use env command**
```bash
env $(cat .env | grep -v '^#' | xargs) go run cmd/server/main.go
```

**Option 3: Source the file (bash/zsh)**
```bash
set -a
source .env
set +a
go run cmd/server/main.go
```

## Verification

### Check if .env exists
```bash
ls -la .env
```

### Verify environment variables are loaded
```bash
make run
# Should output:
# Starting application...
# MongoDB connection established.
# Using database: authdb, collection: users
# Starting web terminal server on http://localhost:8080
```

### Test MongoDB connection
```bash
# The application will fail immediately if MongoDB connection fails
# Look for: "MongoDB connection established."
```

## Troubleshooting

### Error: "Required environment variables not set"

**Cause:** `.env` file missing or variables not loaded

**Solution:**
1. Verify `.env` file exists: `ls -la .env`
2. Check file contents: `cat .env`
3. Ensure no syntax errors in `.env`
4. Use `make run` instead of `go run` directly

### Error: "Failed to connect to MongoDB"

**Cause:** Invalid `DB_DSN` or MongoDB not accessible

**Solution:**
1. Verify MongoDB is running (if local)
2. Check connection string format
3. Test connection: `mongosh "your-connection-string"`
4. For Atlas: Check IP whitelist

### Error: "Failed to exchange code" (OAuth)

**Cause:** Invalid Google OAuth credentials

**Solution:**
1. Verify `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`
2. Check redirect URI in Google Console matches `APP_BASE_URL/auth/google/callback`
3. Ensure Google+ API is enabled

### Error: "bind: address already in use"

**Cause:** Port 8080 already in use

**Solution:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change port in .env
SERVER_PORT=8081
```

## Security Best Practices

### 1. Never Commit .env to Git

The `.gitignore` already includes `.env`:
```gitignore
# env file
.env
```

### 2. Use Strong Session Keys

```bash
# Generate a new key
openssl rand -base64 32
```

### 3. Rotate Credentials Regularly

- Change session key periodically
- Rotate MongoDB passwords
- Update OAuth credentials if compromised

### 4. Use Environment-Specific Files

**Development:**
```bash
.env.development
```

**Production:**
```bash
.env.production
```

**Load specific environment:**
```bash
cp .env.production .env
make run
```

### 5. Restrict MongoDB Access

- Use strong passwords
- Enable IP whitelisting (Atlas)
- Use least-privilege user accounts
- Enable authentication

## Production Deployment

### Using Environment Variables (Recommended)

Instead of `.env` file, set variables directly:

**systemd service:**
```ini
[Service]
Environment="GOOGLE_CLIENT_ID=..."
Environment="GOOGLE_CLIENT_SECRET=..."
Environment="DB_DSN=..."
Environment="SESSION_KEY=..."
Environment="APP_BASE_URL=https://yourdomain.com"
```

**Docker:**
```bash
docker run -e GOOGLE_CLIENT_ID=... -e GOOGLE_CLIENT_SECRET=... ...
```

**Kubernetes:**
```yaml
env:
  - name: GOOGLE_CLIENT_ID
    valueFrom:
      secretKeyRef:
        name: app-secrets
        key: google-client-id
```

### Using .env File

If using `.env` in production:
1. Set restrictive permissions: `chmod 600 .env`
2. Ensure file is not in web-accessible directory
3. Use secrets management (Vault, AWS Secrets Manager, etc.)

## Summary

✅ Create `.env` file with all required variables  
✅ Use `make run` to automatically load environment  
✅ Never commit `.env` to version control  
✅ Use strong, random session keys  
✅ Verify MongoDB connection string format  
✅ Test OAuth credentials in Google Console  

---

**Need help?** Check [QUICKSTART.md](QUICKSTART.md) for a 5-minute setup guide.
