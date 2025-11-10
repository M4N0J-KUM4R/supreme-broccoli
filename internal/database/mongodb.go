package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"supreme-broccoli/internal/models"
)

// MongoDB holds the database connection and collections
type MongoDB struct {
	Client          *mongo.Client
	Database        *mongo.Database
	UsersCollection *mongo.Collection
}

// Connect establishes a connection to MongoDB
func Connect(uri string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	database := client.Database("authdb")
	usersCollection := database.Collection("users")

	log.Println("MongoDB connection established.")
	log.Printf("Using database: %s, collection: %s", database.Name(), usersCollection.Name())

	return &MongoDB{
		Client:          client,
		Database:        database,
		UsersCollection: usersCollection,
	}, nil
}

// Close disconnects from MongoDB
func (db *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("error disconnecting from MongoDB: %v", err)
	}

	log.Println("MongoDB connection closed.")
	return nil
}

// SaveUser saves or updates a user's tokens in the database
func (db *MongoDB) SaveUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	result, err := db.UsersCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("user already exists: %s", user.Email)
		}
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("database operation timeout for user: %s", user.Email)
		}
		return fmt.Errorf("failed to save user %s: %v", user.Email, err)
	}

	if result.UpsertedCount > 0 {
		log.Printf("Inserted new user: %s", user.Email)
	} else if result.ModifiedCount > 0 {
		log.Printf("Updated existing user: %s", user.Email)
	}

	return nil
}

// GetUser retrieves a user's details from the database
func (db *MongoDB) GetUser(email string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": email}

	var user models.User
	err := db.UsersCollection.FindOne(ctx, filter).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return models.User{}, fmt.Errorf("user not found: %s", email)
	}

	if ctx.Err() == context.DeadlineExceeded {
		return models.User{}, fmt.Errorf("database operation timeout for user: %s", email)
	}

	if err != nil {
		return models.User{}, fmt.Errorf("failed to retrieve user %s: %v", email, err)
	}

	return user, nil
}
