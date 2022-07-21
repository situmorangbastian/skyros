package productservice

import "fmt"

// ErrorNotFound represents a custom error for a not found things.
type ErrorNotFound string

// Error retuns error message string.
func (e ErrorNotFound) Error() string {
	return string(e)
}

// ErrorNotFoundf constructs ErrorNotFound with formatted message.
func ErrorNotFoundf(format string, a ...interface{}) ErrorNotFound {
	return ErrorNotFound(fmt.Sprintf(format, a...))
}

// ConflictError represents a custom error for a conflict things.
type ConflictError string

// Error retuns error message string.
func (e ConflictError) Error() string {
	return string(e)
}

// ConflictErrorf constructs ConflictError with formatted message.
func ConflictErrorf(format string, a ...interface{}) ConflictError {
	return ConflictError(fmt.Sprintf(format, a...))
}

// ConstraintError represents a custom error for a contstraint things.
type ConstraintError string

// Error retuns error message string.
func (e ConstraintError) Error() string {
	return string(e)
}

// ConstraintErrorf constructs ConstraintErrorf with formatted message.
func ConstraintErrorf(format string, a ...interface{}) ConstraintError {
	return ConstraintError(fmt.Sprintf(format, a...))
}
