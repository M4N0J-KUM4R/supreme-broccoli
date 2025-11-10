package models

import "time"

// ContactMessage represents a contact form submission
type ContactMessage struct {
	ID        string    `bson:"_id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	Email     string    `bson:"email" json:"email"`
	Subject   string    `bson:"subject" json:"subject"`
	Message   string    `bson:"message" json:"message"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Status    string    `bson:"status" json:"status"` // "new", "read", "responded"
}
