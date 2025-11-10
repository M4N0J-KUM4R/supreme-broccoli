package models

import "time"

// Course represents a learning course in the application
type Course struct {
	ID          string   `bson:"_id" json:"id"`
	Title       string   `bson:"title" json:"title"`
	Instructor  string   `bson:"instructor" json:"instructor"`
	Duration    string   `bson:"duration" json:"duration"`
	Level       string   `bson:"level" json:"level"`
	Thumbnail   string   `bson:"thumbnail" json:"thumbnail"`
	Description string   `bson:"description" json:"description"`
}

// UserProgress tracks a user's progress through a course
type UserProgress struct {
	UserEmail  string    `bson:"user_email" json:"user_email"`
	CourseID   string    `bson:"course_id" json:"course_id"`
	Progress   int       `bson:"progress" json:"progress"` // 0-100
	Enrolled   bool      `bson:"enrolled" json:"enrolled"`
	LastAccess time.Time `bson:"last_access" json:"last_access"`
}

// GetMockCourses returns a slice of sample courses for development and testing
func GetMockCourses() []Course {
	return []Course{
		{
			ID:          "gcp-fundamentals",
			Title:       "Google Cloud Platform Fundamentals",
			Instructor:  "Sarah Chen",
			Duration:    "4 hours",
			Level:       "Beginner",
			Thumbnail:   "/static/images/course-thumbnails/gcp-fundamentals.jpg",
			Description: "Learn the basics of Google Cloud Platform including Compute Engine, Cloud Storage, and networking fundamentals.",
		},
		{
			ID:          "kubernetes-essentials",
			Title:       "Kubernetes Essentials",
			Instructor:  "Michael Rodriguez",
			Duration:    "6 hours",
			Level:       "Intermediate",
			Thumbnail:   "/static/images/course-thumbnails/kubernetes-essentials.jpg",
			Description: "Master container orchestration with Kubernetes. Learn pods, deployments, services, and best practices.",
		},
		{
			ID:          "cloud-shell-mastery",
			Title:       "Cloud Shell Mastery",
			Instructor:  "Emily Watson",
			Duration:    "3 hours",
			Level:       "Beginner",
			Thumbnail:   "/static/images/course-thumbnails/cloud-shell-mastery.jpg",
			Description: "Become proficient with Google Cloud Shell. Learn command-line tools, scripting, and productivity tips.",
		},
		{
			ID:          "terraform-infrastructure",
			Title:       "Infrastructure as Code with Terraform",
			Instructor:  "David Kim",
			Duration:    "8 hours",
			Level:       "Advanced",
			Thumbnail:   "/static/images/course-thumbnails/terraform-infrastructure.jpg",
			Description: "Build and manage cloud infrastructure using Terraform. Learn modules, state management, and best practices.",
		},
		{
			ID:          "docker-containers",
			Title:       "Docker Containers Deep Dive",
			Instructor:  "Lisa Anderson",
			Duration:    "5 hours",
			Level:       "Intermediate",
			Thumbnail:   "/static/images/course-thumbnails/docker-containers.jpg",
			Description: "Master Docker containerization. Learn images, volumes, networking, and multi-stage builds.",
		},
		{
			ID:          "cloud-security",
			Title:       "Cloud Security Best Practices",
			Instructor:  "James Thompson",
			Duration:    "7 hours",
			Level:       "Advanced",
			Thumbnail:   "/static/images/course-thumbnails/cloud-security.jpg",
			Description: "Secure your cloud infrastructure. Learn IAM, encryption, network security, and compliance.",
		},
	}
}
