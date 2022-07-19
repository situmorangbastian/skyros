package userservice

import "fmt"

type ErrorNotFound string

func (e ErrorNotFound) Error() string {
	return string(e)
}

func ErrorNotFoundf(format string, a ...interface{}) ErrorNotFound {
	return ErrorNotFound(fmt.Sprintf(format, a...))
}

type ConflictError string

func (e ConflictError) Error() string {
	return string(e)
}

func ConflictErrorf(format string, a ...interface{}) ConflictError {
	return ConflictError(fmt.Sprintf(format, a...))
}

type ConstraintError string

func (e ConstraintError) Error() string {
	return string(e)
}

func ConstraintErrorf(format string, a ...interface{}) ConstraintError {
	return ConstraintError(fmt.Sprintf(format, a...))
}
