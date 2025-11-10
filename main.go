package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("⚠️  This file has been deprecated!")
	fmt.Println("")
	fmt.Println("The application has been refactored into a modular structure.")
	fmt.Println("")
	fmt.Println("To build and run the application, use:")
	fmt.Println("  make build    # Build the application")
	fmt.Println("  make run      # Run the application")
	fmt.Println("")
	fmt.Println("Or directly:")
	fmt.Println("  go build -o bin/supreme-broccoli cmd/server/main.go")
	fmt.Println("  ./bin/supreme-broccoli")
	fmt.Println("")
	fmt.Println("See PROJECT_STRUCTURE.md for details about the new structure.")
	fmt.Println("The old main.go has been backed up to: backup/main.go.old")
	fmt.Println("")
	os.Exit(1)
}
