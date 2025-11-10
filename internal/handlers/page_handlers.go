package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"supreme-broccoli/internal/helpers"
	"supreme-broccoli/internal/models"

	"github.com/gorilla/sessions"
)

// PageHandlers handles page rendering
type PageHandlers struct {
	SessionStore *sessions.CookieStore
	templates    *template.Template
}

// NewPageHandlers creates a new PageHandlers instance
func NewPageHandlers(sessionStore *sessions.CookieStore) *PageHandlers {
	// Parse all templates
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Printf("Warning: Failed to parse templates: %v", err)
	}
	
	// Parse partials
	templates, err = templates.ParseGlob("templates/partials/*.html")
	if err != nil {
		log.Printf("Warning: Failed to parse partial templates: %v", err)
	}

	return &PageHandlers{
		SessionStore: sessionStore,
		templates:    templates,
	}
}

// HandleHome renders the home page
func (h *PageHandlers) HandleHome(w http.ResponseWriter, r *http.Request) {
	// Get page data from session
	pageData := helpers.GetPageData(r, h.SessionStore, "home")

	// Render the home template
	err := h.templates.ExecuteTemplate(w, "home.html", pageData)
	if err != nil {
		log.Printf("Error rendering home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleCourses renders the courses page with available courses
func (h *PageHandlers) HandleCourses(w http.ResponseWriter, r *http.Request) {
	// Get page data from session
	pageData := helpers.GetPageData(r, h.SessionStore, "courses")

	// Fetch mock courses with progress
	courses := helpers.GetCoursesWithProgress(pageData.User)

	// Create courses page data
	coursesData := helpers.CoursesPageData{
		PageData: *pageData,
		Courses:  courses,
	}

	// Render the courses template
	err := h.templates.ExecuteTemplate(w, "courses.html", coursesData)
	if err != nil {
		log.Printf("Error rendering courses template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}


// HandleProfile renders the user profile page
func (h *PageHandlers) HandleProfile(w http.ResponseWriter, r *http.Request) {
	// Get page data from session
	pageData := helpers.GetPageData(r, h.SessionStore, "profile")

	// Get profile data with statistics and activity
	profileData := helpers.GetProfileData(pageData.User)
	profileData.PageData = *pageData

	// Render the profile template
	err := h.templates.ExecuteTemplate(w, "profile.html", profileData)
	if err != nil {
		log.Printf("Error rendering profile template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}


// HandleSettingsPage renders the settings page (GET)
func (h *PageHandlers) HandleSettingsPage(w http.ResponseWriter, r *http.Request) {
	// Get page data from session
	pageData := helpers.GetPageData(r, h.SessionStore, "settings")

	// Get success/error messages from session if any
	session, _ := h.SessionStore.Get(r, "auth-session")
	successMsg, _ := session.Values["success_message"].(string)
	errorMsg, _ := session.Values["error_message"].(string)
	
	// Clear messages from session
	delete(session.Values, "success_message")
	delete(session.Values, "error_message")
	session.Save(r, w)

	// Get user settings (use defaults if not set)
	settings := pageData.User.Settings
	if settings.TerminalFontSize == 0 {
		settings = models.DefaultSettings()
	}

	// Create settings page data
	settingsData := helpers.SettingsPageData{
		PageData:       *pageData,
		Settings:       settings,
		SuccessMessage: successMsg,
		ErrorMessage:   errorMsg,
	}

	// Render the settings template
	err := h.templates.ExecuteTemplate(w, "settings.html", settingsData)
	if err != nil {
		log.Printf("Error rendering settings template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleSettingsUpdate processes settings form submission (POST)
func (h *PageHandlers) HandleSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		h.setSessionMessage(r, w, "", "Invalid form data")
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	// Get current user from session
	session, _ := h.SessionStore.Get(r, "auth-session")
	email, ok := session.Values["email"].(string)
	if !ok || email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse and validate settings
	_, validationErr := h.parseAndValidateSettings(r)
	if validationErr != "" {
		h.setSessionMessage(r, w, "", validationErr)
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	// TODO: Update settings in database
	// For now, we'll just show a success message
	// In a real implementation, you would call a database function here
	// settings would be used here: db.UpdateUserSettings(email, settings)

	h.setSessionMessage(r, w, "Settings saved successfully!", "")
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// parseAndValidateSettings parses and validates settings from form data
func (h *PageHandlers) parseAndValidateSettings(r *http.Request) (models.UserSettings, string) {
	settings := models.UserSettings{}

	// Parse terminal font size
	fontSize := r.FormValue("terminalFontSize")
	if fontSize != "" {
		var size int
		_, err := fmt.Sscanf(fontSize, "%d", &size)
		if err != nil || size < 10 || size > 24 {
			return settings, "Font size must be between 10 and 24"
		}
		settings.TerminalFontSize = size
	} else {
		settings.TerminalFontSize = 14
	}

	// Parse color scheme
	colorScheme := r.FormValue("terminalColorScheme")
	validSchemes := map[string]bool{"dark": true, "light": true, "solarized": true, "monokai": true}
	if !validSchemes[colorScheme] {
		return settings, "Invalid color scheme"
	}
	settings.TerminalColorScheme = colorScheme

	// Parse cursor style
	cursorStyle := r.FormValue("terminalCursorStyle")
	validCursors := map[string]bool{"block": true, "underline": true, "bar": true}
	if !validCursors[cursorStyle] {
		return settings, "Invalid cursor style"
	}
	settings.TerminalCursorStyle = cursorStyle

	// Parse checkboxes
	settings.EmailNotifications = r.FormValue("emailNotifications") == "on"
	settings.CourseUpdates = r.FormValue("courseUpdates") == "on"

	return settings, ""
}

// setSessionMessage sets a success or error message in the session
func (h *PageHandlers) setSessionMessage(r *http.Request, w http.ResponseWriter, success, error string) {
	session, _ := h.SessionStore.Get(r, "auth-session")
	if success != "" {
		session.Values["success_message"] = success
	}
	if error != "" {
		session.Values["error_message"] = error
	}
	session.Save(r, w)
}


// HandleAbout renders the about page
func (h *PageHandlers) HandleAbout(w http.ResponseWriter, r *http.Request) {
	// Get page data from session
	pageData := helpers.GetPageData(r, h.SessionStore, "about")

	// Render the about template
	err := h.templates.ExecuteTemplate(w, "about.html", pageData)
	if err != nil {
		log.Printf("Error rendering about template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}


// HandleContactPage renders the contact page (GET)
func (h *PageHandlers) HandleContactPage(w http.ResponseWriter, r *http.Request) {
	// Get page data from session
	pageData := helpers.GetPageData(r, h.SessionStore, "contact")

	// Get success/error messages from session if any
	session, _ := h.SessionStore.Get(r, "auth-session")
	successMsg, _ := session.Values["success_message"].(string)
	errorMsg, _ := session.Values["error_message"].(string)
	
	// Clear messages from session
	delete(session.Values, "success_message")
	delete(session.Values, "error_message")
	session.Save(r, w)

	// Create contact page data
	contactData := helpers.ContactPageData{
		PageData:       *pageData,
		SuccessMessage: successMsg,
		ErrorMessage:   errorMsg,
	}

	// Render the contact template
	err := h.templates.ExecuteTemplate(w, "contact.html", contactData)
	if err != nil {
		log.Printf("Error rendering contact template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleContactSubmit processes contact form submission (POST)
func (h *PageHandlers) HandleContactSubmit(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		h.setSessionMessage(r, w, "", "Invalid form data")
		http.Redirect(w, r, "/contact", http.StatusSeeOther)
		return
	}

	// Get form values
	name := r.FormValue("name")
	email := r.FormValue("email")
	subject := r.FormValue("subject")
	message := r.FormValue("message")

	// Validate form data
	if validationErr := h.validateContactForm(name, email, subject, message); validationErr != "" {
		h.setSessionMessage(r, w, "", validationErr)
		http.Redirect(w, r, "/contact", http.StatusSeeOther)
		return
	}

	// TODO: Save contact message to database
	// For now, we'll just log it and show a success message
	log.Printf("Contact form submission - Name: %s, Email: %s, Subject: %s", name, email, subject)

	h.setSessionMessage(r, w, "Thank you for contacting us! We'll get back to you soon.", "")
	http.Redirect(w, r, "/contact", http.StatusSeeOther)
}

// validateContactForm validates contact form fields
func (h *PageHandlers) validateContactForm(name, email, subject, message string) string {
	// Validate name
	if name == "" {
		return "Name is required"
	}
	if len(name) < 2 {
		return "Name must be at least 2 characters"
	}

	// Validate email
	if email == "" {
		return "Email is required"
	}
	// Simple email validation
	if !h.isValidEmail(email) {
		return "Please enter a valid email address"
	}

	// Validate subject
	if subject == "" {
		return "Subject is required"
	}
	if len(subject) < 3 {
		return "Subject must be at least 3 characters"
	}

	// Validate message
	if message == "" {
		return "Message is required"
	}
	if len(message) < 10 {
		return "Message must be at least 10 characters"
	}

	return ""
}

// isValidEmail performs basic email validation
func (h *PageHandlers) isValidEmail(email string) bool {
	// Simple email validation - check for @ and .
	atIndex := -1
	dotIndex := -1
	
	for i, c := range email {
		if c == '@' {
			if atIndex != -1 {
				return false // Multiple @ symbols
			}
			atIndex = i
		}
		if c == '.' && atIndex != -1 {
			dotIndex = i
		}
	}
	
	return atIndex > 0 && dotIndex > atIndex+1 && dotIndex < len(email)-1
}
