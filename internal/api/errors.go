// API error handling
package api

import (
	"encoding/json"
	"net/http"
)

// Error represents an API error
type Error struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Status  int    `json:"status"`
	Details any    `json:"details,omitempty"`
}

// NewError creates a new API error
func NewError(message, code string, status int) *Error {
	return &Error{
		Message: message,
		Code:    code,
		Status:  status,
	}
}

// WithDetails adds details to the error
func (e *Error) WithDetails(details any) *Error {
	e.Details = details
	return e
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, err *Error) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")
	
	// Set status code
	w.WriteHeader(err.Status)
	
	// Write JSON response
	json.NewEncoder(w).Encode(err)
}