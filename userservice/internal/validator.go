package internal

import (
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/situmorangbastian/skyros/userservice"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() CustomValidator {
	cv := validator.New()

	return CustomValidator{
		validator: cv,
	}
}

func (cv CustomValidator) Validate(data interface{}) error {
	if err := cv.validator.Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.ActualTag() {
			case "email":
				return userservice.ConstraintError("invalid email")
			default:
				return userservice.ConstraintError(strings.ToLower(err.Field()) + " " + err.ActualTag())
			}
		}
	}

	return nil
}
