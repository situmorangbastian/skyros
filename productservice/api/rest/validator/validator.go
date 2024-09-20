package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"

	internalErr "github.com/situmorangbastian/skyros/productservice/internal/error"
)

// CustomValidator is struct for custom validator
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator is function to init custom validator
func NewValidator() CustomValidator {
	cv := validator.New()

	return CustomValidator{
		validator: cv,
	}
}

// Validate is method implementation for validating struct
func (cv CustomValidator) Validate(data interface{}) error {
	if err := cv.validator.Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.ActualTag() {
			case "email":
				return internalErr.ConstraintError("invalid email")
			default:
				return internalErr.ConstraintError(strings.ToLower(err.Field()) + " " + err.ActualTag())
			}
		}
	}

	return nil
}
