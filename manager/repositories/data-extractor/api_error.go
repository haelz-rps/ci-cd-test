package dataextractor

import "fmt"

type ApiError struct {
	Code    int
	Message string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("API failed with status: %d and body '%s'", e.Code, e.Message)
}

func NewApiError(statusCode int, body []byte) *ApiError {
	return &ApiError{statusCode, string(body)}
}
