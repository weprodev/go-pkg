// Package httperr provides structured HTTP error types for Go web services.
//
// It defines ServiceError, a type that implements the error interface while
// carrying an HTTP status code and optional detail messages. Sentinel errors
// for common HTTP failure modes are pre-defined for convenience.
//
// This package has zero external dependencies — it relies only on the
// standard library — so any package can import it without pulling in
// framework-specific transitive dependencies.
package httperr

import (
	"errors"
	"fmt"
	"net/http"
)

// HTTP status codes for common error responses.
const (
	StatusBadRequest     = http.StatusBadRequest          // 400
	StatusUnauthorized   = http.StatusUnauthorized        // 401
	StatusNotFound       = http.StatusNotFound            // 404
	StatusConflict       = http.StatusConflict            // 409
	StatusUnprocessable  = http.StatusUnprocessableEntity // 422
	StatusInternalServer = http.StatusInternalServerError // 500
)

// ServiceError represents a service error with an HTTP status code and messages.
type ServiceError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

// Error implements the error interface.
func (e ServiceError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("%s: %v", e.Message, e.Errors)
	}
	return e.Message
}

// Is implements errors.Is for proper error comparison.
// Two ServiceErrors are considered equal if they share the same Code and Message.
func (e ServiceError) Is(target error) bool {
	t, ok := target.(ServiceError)
	if !ok {
		return false
	}
	return e.Code == t.Code && e.Message == t.Message
}

// NewServiceError creates a new ServiceError.
func NewServiceError(code int, message string, details ...string) ServiceError {
	return ServiceError{
		Code:    code,
		Message: message,
		Errors:  details,
	}
}

// Common sentinel errors reusable across any service.
var (
	ErrBadRequest         = NewServiceError(StatusBadRequest, "bad request")
	ErrUnauthorized       = NewServiceError(StatusUnauthorized, "unauthorized access")
	ErrNotFound           = NewServiceError(StatusNotFound, "resource not found")
	ErrConflict           = NewServiceError(StatusConflict, "resource conflict")
	ErrValidation         = NewServiceError(StatusUnprocessable, "validation error")
	ErrInvalidRequestBody = NewServiceError(StatusBadRequest, "invalid request body")
	ErrInternalServer     = NewServiceError(StatusInternalServer, "internal server error")
)

// AsServiceError safely converts an error to *ServiceError using errors.As.
// Returns nil if the error is nil or not a ServiceError.
func AsServiceError(err error) *ServiceError {
	if err == nil {
		return nil
	}

	var se ServiceError
	if errors.As(err, &se) {
		return &se
	}
	return nil
}
