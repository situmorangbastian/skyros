package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() CustomValidator {
	return CustomValidator{
		validator: validator.New(),
	}
}

func (cv CustomValidator) Validate(data interface{}) error {
	if err := cv.validator.Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
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
