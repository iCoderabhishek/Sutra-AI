package api

import "fmt"

// n error returned by the Node.js backend.
type APIError struct {
	StatusCode int    `json:"-"`
	Method     string `json:"-"`
	URL        string `json:"-"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error (%s %s) - Status: %d, Message: %s", e.Method, e.URL, e.StatusCode, e.Message)
}
