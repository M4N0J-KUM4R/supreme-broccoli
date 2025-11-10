package models

import "time"

// User represents a user in the system with OAuth tokens
type User struct {
	Email        string       `bson:"_id"`
	AccessToken  string       `bson:"access_token"`
	RefreshToken string       `bson:"refresh_token"`
	TokenExpiry  time.Time    `bson:"token_expiry"`
	Role         string       `bson:"role"`
	DisplayName  string       `bson:"display_name"`
	ProfilePic   string       `bson:"profile_pic"`
	MemberSince  time.Time    `bson:"member_since"`
	Settings     UserSettings `bson:"settings"`
}
