package errors

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid request input")
	ErrUnauthorized = errors.New("unauthorized")
)

type BusinessError struct {
	Code    int
	Message string
	Err     error
}

func NewBusinessError(code int, message string, cause error) *BusinessError {
	return &BusinessError{Code: code, Message: message, Err: cause}
}

func (e *BusinessError) Error() string { return e.Message }
func (e *BusinessError) Unwrap() error { return e.Err }
