package models

// UserSettings represents user preferences and configuration
type UserSettings struct {
	TerminalFontSize      int    `bson:"terminal_font_size" json:"terminal_font_size"`
	TerminalColorScheme   string `bson:"terminal_color_scheme" json:"terminal_color_scheme"`
	TerminalCursorStyle   string `bson:"terminal_cursor_style" json:"terminal_cursor_style"`
	EmailNotifications    bool   `bson:"email_notifications" json:"email_notifications"`
	CourseUpdates         bool   `bson:"course_updates" json:"course_updates"`
}

// DefaultSettings returns default user settings
func DefaultSettings() UserSettings {
	return UserSettings{
		TerminalFontSize:      14,
		TerminalColorScheme:   "dark",
		TerminalCursorStyle:   "block",
		EmailNotifications:    true,
		CourseUpdates:         true,
	}
}
