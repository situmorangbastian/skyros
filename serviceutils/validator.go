package serviceutils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() CustomValidator {
	return CustomValidator{
		validator: validator.New(),
	}
}

func (cv CustomValidator) Validate(data interface{}) error {
	if err := cv.validator.Struct(data); err != nil {
		var validationErrs validator.ValidationErrors
		if !errors.As(err, &validationErrs) {
			return status.Error(codes.InvalidArgument, "invalid request")
		}
		for _, err := range validationErrs {
			switch err.ActualTag() {
			case "email":
				return status.Error(codes.InvalidArgument, "invalid email")
			default:
				return status.Error(codes.InvalidArgument, strings.ToLower(fmt.Sprintf("%s %s", err.Field(), err.ActualTag())))
			}
		}
	}

	return nil
}
