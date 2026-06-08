package models

// ErrorResponse is the standard error body for all HTTP errors.
type ErrorResponse struct {
	Error string `json:"error"`
}
