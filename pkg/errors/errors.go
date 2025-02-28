package errors

import (
	"fmt"
)

// AppError represents an application error
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error returns the error message
func (e AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code int, message string, err error) AppError {
	return AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string, err error) AppError {
	return NewAppError(400, message, err)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, err error) AppError {
	return NewAppError(404, message, err)
}

// NewInternalServerError creates a new internal server error
func NewInternalServerError(message string, err error) AppError {
	return NewAppError(500, message, err)
}