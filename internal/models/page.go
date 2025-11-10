package models

// PageData represents common data passed to page templates
type PageData struct {
	IsAuthenticated bool
	ActivePage      string
	User            *User
}
