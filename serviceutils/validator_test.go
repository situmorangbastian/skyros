package serviceutils

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Regression: a non-struct argument made validator return InvalidValidationError,
// and the unchecked type assertion panicked.
func TestValidate_NonStructDoesNotPanic(t *testing.T) {
	err := NewCustomValidator().Validate("not a struct")
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got %v, want InvalidArgument", err)
	}
}
