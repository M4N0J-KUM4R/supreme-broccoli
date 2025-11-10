package helpers

import (
	"net/http"

	"supreme-broccoli/internal/models"

	"github.com/gorilla/sessions"
)

// PageData is an alias for models.PageData for convenience
type PageData = models.PageData

// CoursesPageData extends PageData with course-specific data
type CoursesPageData struct {
	PageData
	Courses []CourseWithProgress
}

// CourseWithProgress combines Course data with user progress
type CourseWithProgress struct {
	models.Course
	Progress int
	Enrolled bool
}

// ProfilePageData extends PageData with profile-specific data
type ProfilePageData struct {
	PageData
	Stats          ProfileStats
	RecentActivity []ActivityItem
}

// ProfileStats contains user statistics
type ProfileStats struct {
	EnrolledCourses  int
	CompletedCourses int
	LearningHours    int
	CurrentStreak    int
}

// ActivityItem represents a recent activity entry
type ActivityItem struct {
	Type        string // "course", "completion", "login"
	Title       string
	Description string
	TimeAgo     string
}

// SettingsPageData extends PageData with settings-specific data
type SettingsPageData struct {
	PageData
	Settings       models.UserSettings
	SuccessMessage string
	ErrorMessage   string
}

// ContactPageData extends PageData with contact-specific data
type ContactPageData struct {
	PageData
	SuccessMessage string
	ErrorMessage   string
}

// GetPageData creates a PageData struct from the current session
func GetPageData(r *http.Request, sessionStore *sessions.CookieStore, activePage string) *models.PageData {
	session, _ := sessionStore.Get(r, "auth-session")
	
	pageData := &models.PageData{
		IsAuthenticated: false,
		ActivePage:      activePage,
		User:            nil,
	}

	// Check if user is authenticated
	if email, ok := session.Values["email"].(string); ok && email != "" {
		pageData.IsAuthenticated = true
		
		// Create a basic user object from session data
		user := &models.User{
			Email: email,
		}
		
		// Add role if available in session
		if role, ok := session.Values["role"].(string); ok {
			user.Role = role
		}
		
		pageData.User = user
	}

	return pageData
}

// GetCoursesWithProgress returns courses with user progress information
// For now, this returns mock data with sample progress for demonstration
func GetCoursesWithProgress(user *models.User) []CourseWithProgress {
	// Get mock courses
	courses := models.GetMockCourses()
	
	// Convert to CourseWithProgress with mock progress data
	coursesWithProgress := make([]CourseWithProgress, len(courses))
	
	for i, course := range courses {
		// Mock progress data - in a real app, this would come from the database
		enrolled := false
		progress := 0
		
		// For demonstration, mark first 2 courses as enrolled with progress
		if user != nil && i < 2 {
			enrolled = true
			if i == 0 {
				progress = 65 // First course 65% complete
			} else {
				progress = 30 // Second course 30% complete
			}
		}
		
		coursesWithProgress[i] = CourseWithProgress{
			Course:   course,
			Progress: progress,
			Enrolled: enrolled,
		}
	}
	
	return coursesWithProgress
}


// GetProfileData returns profile page data with mock statistics and activity
func GetProfileData(user *models.User) ProfilePageData {
	// Mock statistics - in a real app, this would come from the database
	stats := ProfileStats{
		EnrolledCourses:  2,
		CompletedCourses: 0,
		LearningHours:    12,
		CurrentStreak:    5,
	}

	// Mock recent activity - in a real app, this would come from the database
	recentActivity := []ActivityItem{
		{
			Type:        "course",
			Title:       "Started Kubernetes Essentials",
			Description: "Began learning container orchestration",
			TimeAgo:     "2 hours ago",
		},
		{
			Type:        "course",
			Title:       "Continued GCP Fundamentals",
			Description: "Completed module 3: Cloud Storage",
			TimeAgo:     "1 day ago",
		},
		{
			Type:        "course",
			Title:       "Enrolled in GCP Fundamentals",
			Description: "Started your cloud learning journey",
			TimeAgo:     "3 days ago",
		},
	}

	return ProfilePageData{
		Stats:          stats,
		RecentActivity: recentActivity,
	}
}
