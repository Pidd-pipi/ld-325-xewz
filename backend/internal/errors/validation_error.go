package errors

type ValidationError struct {
	Message string
	Err     error
}

func (e *ValidationError) Error() string { return e.Message }
func (e *ValidationError) Unwrap() error { return e.Err }

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message, Err: ErrInvalidInput}
}
