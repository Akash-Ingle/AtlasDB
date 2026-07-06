package models

import "fmt"

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func NewAPIError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

var (
	ErrBadRequest          = &APIError{Code: "bad_request", Message: "The request body is invalid"}
	ErrUnauthorized        = &APIError{Code: "unauthorized", Message: "Authentication required"}
	ErrForbidden           = &APIError{Code: "forbidden", Message: "Insufficient permissions"}
	ErrNotFound            = &APIError{Code: "not_found", Message: "Resource not found"}
	ErrRateLimited         = &APIError{Code: "rate_limited", Message: "Too many requests"}
	ErrInternalServer      = &APIError{Code: "internal_error", Message: "An internal error occurred"}
	ErrServiceUnavailable  = &APIError{Code: "service_unavailable", Message: "Service temporarily unavailable"}
)
