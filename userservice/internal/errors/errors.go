package errors

import (
	"fmt"
)

type ConstraintError string

func (e ConstraintError) Error() string {
	return string(e)
}

func ConstraintErrorf(format string, a ...interface{}) ConstraintError {
	return ConstraintError(fmt.Sprintf(format, a...))
}

type NotFoundError string

func (e NotFoundError) Error() string {
	return string(e)
}

func NotFoundErrorf(format string, a ...interface{}) NotFoundError {
	return NotFoundError(fmt.Sprintf(format, a...))
}

type ConflictError string

func (e ConflictError) Error() string {
	return string(e)
}

func ConflictErrorf(format string, a ...interface{}) ConflictError {
	return ConflictError(fmt.Sprintf(format, a...))
}
