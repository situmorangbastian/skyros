package models

import (
	"testing"

	"github.com/situmorangbastian/skyros/serviceutils"
)

// Regression: without `dive`, per-item rules were never evaluated.
func TestOrderValidation_ValidatesItems(t *testing.T) {
	v := serviceutils.NewCustomValidator()
	base := Order{DestinationAddress: "somewhere"}

	bad := base
	bad.Items = []OrderProduct{{ProductID: "p1", Quantity: -5}}
	if err := v.Validate(bad); err == nil {
		t.Error("negative quantity should be rejected")
	}

	missing := base
	missing.Items = []OrderProduct{{Quantity: 1}}
	if err := v.Validate(missing); err == nil {
		t.Error("missing product id should be rejected")
	}

	ok := base
	ok.Items = []OrderProduct{{ProductID: "p1", Quantity: 2}}
	if err := v.Validate(ok); err != nil {
		t.Errorf("valid order rejected: %v", err)
	}
}
