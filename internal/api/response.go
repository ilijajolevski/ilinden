// API response handling
package api

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// NewResponse creates a new API response
func NewResponse(success bool, message string, data any) *Response {
	return &Response{
		Success: success,
		Message: message,
		Data:    data,
	}
}

// WriteResponse writes a response to the HTTP writer
func WriteResponse(w http.ResponseWriter, status int, resp *Response) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")
	
	// Set status code
	w.WriteHeader(status)
	
	// Write JSON response
	json.NewEncoder(w).Encode(resp)
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data any) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")
	
	// Set status code
	w.WriteHeader(status)
	
	// Write JSON response
	json.NewEncoder(w).Encode(data)
}