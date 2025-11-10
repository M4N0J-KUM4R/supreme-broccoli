package handlers

import (
	"fmt"
	"net/http"
)

// HandleAdmin serves the admin-only page
func HandleAdmin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the Admin-Only Page!")
}
