package apperrors

import "fmt"

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Message)
}

func NewAPIError(statusCode int, msg string) *APIError {
	return &APIError{StatusCode: statusCode, Message: msg}
}
