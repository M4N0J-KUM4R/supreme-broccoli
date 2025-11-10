# Quick Start Guide

Get supreme-broccoli up and running in 5 minutes!

## Prerequisites

- Go 1.24+
- MongoDB (local or Atlas)
- Google OAuth credentials

## Step 1: Clone and Setup

```bash
cd supreme-broccoli
```

## Step 2: Configure Environment

Create `.env` file:

```bash
# Google OAuth (get from Google Cloud Console)
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret

# MongoDB Connection
# Option A: Local MongoDB
DB_DSN=mongodb://localhost:27017/authdb

# Option B: MongoDB Atlas (recommended)
DB_DSN=mongodb+srv://user:pass@cluster.mongodb.net/authdb?retryWrites=true&w=majority

# Session Key (generate random string)
SESSION_KEY=$(openssl rand -base64 32)

# Application URL
APP_BASE_URL=http://localhost:8080
```

## Step 3: Install Dependencies

```bash
make deps
```

## Step 4: Build

```bash
make build
```

## Step 5: Run

```bash
make run
# The .env file is automatically loaded!
```

You should see:
```
Starting application...
MongoDB connection established.
Using database: authdb, collection: users
Starting web terminal server on http://localhost:8080
```

## Step 6: Access

Open browser: http://localhost:8080

## Common Commands

```bash
# Development
make run          # Run application
make build        # Build binary
make clean        # Clean artifacts
make fmt          # Format code

# Migration (if coming from MySQL)
make migrate-run  # Migrate data from MySQL to MongoDB

# Help
make help         # Show all commands
```

## Troubleshooting

### MongoDB Connection Failed

**Check MongoDB is running:**
```bash
# Local MongoDB
mongosh --eval "db.adminCommand('ping')"

# MongoDB Atlas
# Verify IP whitelist and credentials
```

### Environment Variables Not Set

**Error:** `Required environment variables not set`

**Solution:** Ensure `.env` file exists with all required variables

### Port Already in Use

**Error:** `bind: address already in use`

**Solution:** 
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

## Next Steps

1. **Configure Google OAuth:**
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create OAuth 2.0 credentials
   - Add authorized redirect URI: `http://localhost:8080/auth/google/callback`

2. **Set up MongoDB:**
   - Local: Install MongoDB Community Edition
   - Cloud: Create free MongoDB Atlas cluster

3. **Test the application:**
   - Visit http://localhost:8080
   - Click "Sign in with Google"
   - Access terminal after authentication

4. **Deploy to production:**
   - Build: `make build`
   - Copy `bin/supreme-broccoli` to server
   - Set production environment variables
   - Run with process manager (systemd, pm2, etc.)

## Documentation

- [README.md](README.md) - Full documentation
- [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) - Code organization
- [REFACTORING_SUMMARY.md](REFACTORING_SUMMARY.md) - Recent changes

## Support

For issues or questions:
1. Check documentation
2. Review logs for error messages
3. Verify environment configuration
4. Check MongoDB connectivity

---

**Ready to go!** 🚀
