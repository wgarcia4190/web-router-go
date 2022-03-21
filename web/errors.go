package web

import "github.com/pkg/errors"

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
    Field string `json:"field"`
    Error string `json:"error"`
}

// ErrorResponse how we respond to clients when something goes wrong.
type ErrorResponse struct {
    Error  string       `json:"error"`
    Fields []FieldError `json:"fields,omitempty"`
}

// Error is used to add web information to a request error.
type Error struct {
    Err    error
    Status int
    Fields []FieldError
}

func (e *Error) Error() string {
    return e.Err.Error()
}

// NewRequestError is used when a know error condition is encountered.
func NewRequestError(err error, status int) error {
    return &Error{Err: err, Status: status}
}

// Shutdown is a type used to help with the graceful termination of the service.
type shutdown struct {
    Message string
}

// Error is the implementation of the error interface
func (s *shutdown) Error() string {
    return s.Message
}

// NewShutdownError returns an error that causes the framework to signal
// a graceful shutdown
func NewShutdownError(message string) error {
    return &shutdown{
        Message: message,
    }
}

// IsShutdown checks to see if the shutdown error is contained
// in the specified error value.
func IsShutdown(err error) bool {
    if _, ok := errors.Cause(err).(*shutdown); ok {
        return true
    }
    return false
}
