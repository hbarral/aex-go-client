package aex

import "fmt"

// APIError represents an error returned by the AEX API in the response
// envelope: a result code other than "0" together with a message
// describing the failure.
type APIError struct {
	// Code is the operation result code (codigo). "0" means success.
	Code string
	// Message describes the executed process or the corresponding error.
	Message string
	// Endpoint is the API path that produced the error.
	Endpoint string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("aex: %s: API error %s: %s", e.Endpoint, e.Code, e.Message)
}

// HTTPError represents a non-2xx HTTP status returned by the AEX API.
type HTTPError struct {
	// Endpoint is the API path that produced the error.
	Endpoint string
	// StatusCode is the HTTP status code of the response.
	StatusCode int
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("aex: %s: unexpected HTTP status %d", e.Endpoint, e.StatusCode)
}
