package utils

import (
	"log"
	"net/http"
)

// SafeError logs the detailed error internally and returns a generic error to the client
func SafeError(w http.ResponseWriter, err error, userMessage string, statusCode int) {
	// Log detailed error for debugging (internal only)
	log.Printf("[ERROR] %s: %v", userMessage, err)

	// Send generic error to client (no sensitive details)
	JSONError(w, userMessage, statusCode)
}

// SafeErrorf is like SafeError but with formatted user message
func SafeErrorf(w http.ResponseWriter, err error, userMessageFormat string, statusCode int, args ...interface{}) {
	// Log detailed error for debugging
	log.Printf("[ERROR] %s: %v", userMessageFormat, err)

	// Send formatted generic error to client
	JSONError(w, userMessageFormat, statusCode)
}
