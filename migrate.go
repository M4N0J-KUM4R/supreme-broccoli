package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// User struct matching the main application
type User struct {
	Email        string    `bson:"_id"`
	AccessToken  string    `bson:"access_token"`
	RefreshToken string    `bson:"refresh_token"`
	TokenExpiry  time.Time `bson:"token_expiry"`
	Role         string    `bson:"role"`
}

func main() {
	log.Println("=== Starting MySQL to MongoDB Migration ===")

	// Get connection strings from environment
	mysqlDSN := os.Getenv("MYSQL_DSN")
	mongoDSN := os.Getenv("DB_DSN")

	if mysqlDSN == "" {
		log.Fatal("MYSQL_DSN environment variable not set. Example: admin:password@tcp(host:3306)/authdb?parseTime=true")
	}
	if mongoDSN == "" {
		log.Fatal("DB_DSN environment variable not set. Example: mongodb://admin:password@host:27017/authdb")
	}

	// Connect to MySQL
	log.Println("Connecting to MySQL...")
	mysqlDB, err := connectMySQL(mysqlDSN)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer mysqlDB.Close()
	log.Println("✓ Connected to MySQL")

	// Connect to MongoDB
	log.Println("Connecting to MongoDB...")
	mongoClient, mongoCollection, err := connectMongoDB(mongoDSN)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		mongoClient.Disconnect(ctx)
	}()
	log.Println("✓ Connected to MongoDB")

	// Query all users from MySQL
	log.Println("\nQuerying users from MySQL...")
	users, err := queryMySQLUsers(mysqlDB)
	if err != nil {
		log.Fatalf("Failed to query MySQL users: %v", err)
	}
	log.Printf("✓ Found %d users in MySQL", len(users))

	if len(users) == 0 {
		log.Println("\nNo users to migrate. Exiting.")
		return
	}

	// Migrate users to MongoDB
	log.Println("\nMigrating users to MongoDB...")
	successCount, failCount := migrateUsers(mongoCollection, users)
	log.Printf("✓ Migration complete: %d successful, %d failed", successCount, failCount)

	// Verify migration
	log.Println("\nVerifying migration...")
	if err := verifyMigration(mysqlDB, mongoCollection); err != nil {
		log.Printf("⚠ Verification warning: %v", err)
	} else {
		log.Println("✓ Verification successful: user counts match")
	}

	log.Println("\n=== Migration Complete ===")
}

// connectMySQL establishes connection to MySQL database
func connectMySQL(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}

	return db, nil
}

// connectMongoDB establishes connection to MongoDB database
func connectMongoDB(uri string) (*mongo.Client, *mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Verify connectivity
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	database := client.Database("authdb")
	collection := database.Collection("users")

	return client, collection, nil
}

// queryMySQLUsers retrieves all users from MySQL users table
func queryMySQLUsers(db *sql.DB) ([]User, error) {
	query := `SELECT email, access_token, refresh_token, token_expiry, role FROM users`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %v", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Email, &user.AccessToken, &user.RefreshToken, &user.TokenExpiry, &user.Role); err != nil {
			log.Printf("⚠ Failed to scan user row: %v", err)
			continue
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return users, nil
}

// migrateUsers inserts users into MongoDB collection
func migrateUsers(collection *mongo.Collection, users []User) (int, int) {
	successCount := 0
	failCount := 0

	for _, user := range users {
		if err := insertUser(collection, user); err != nil {
			log.Printf("✗ Failed to migrate user %s: %v", user.Email, err)
			failCount++
		} else {
			log.Printf("✓ Migrated user: %s (role: %s)", user.Email, user.Role)
			successCount++
		}
	}

	return successCount, failCount
}

// insertUser inserts a single user document into MongoDB
func insertUser(collection *mongo.Collection, user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use upsert to handle potential duplicates gracefully
	filter := bson.M{"_id": user.Email}
	update := bson.M{
		"$set": bson.M{
			"access_token":  user.AccessToken,
			"refresh_token": user.RefreshToken,
			"token_expiry":  user.TokenExpiry,
			"role":          user.Role,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert user: %v", err)
	}

	return nil
}

// verifyMigration compares user counts between MySQL and MongoDB
func verifyMigration(mysqlDB *sql.DB, mongoCollection *mongo.Collection) error {
	// Count users in MySQL
	var mysqlCount int
	if err := mysqlDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&mysqlCount); err != nil {
		return fmt.Errorf("failed to count MySQL users: %v", err)
	}

	// Count users in MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mongoCount, err := mongoCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count MongoDB users: %v", err)
	}

	log.Printf("MySQL users: %d", mysqlCount)
	log.Printf("MongoDB users: %d", int(mongoCount))

	if mysqlCount != int(mongoCount) {
		return fmt.Errorf("user count mismatch: MySQL has %d users, MongoDB has %d users", mysqlCount, mongoCount)
	}

	return nil
}
