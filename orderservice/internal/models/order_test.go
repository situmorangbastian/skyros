package models

import (
	"strings"
	"testing"

	"github.com/situmorangbastian/skyros/serviceutils"
)

const validProductID = "0b6f4f7e-3c6a-4d0c-9a51-2f4b8f0e9a11"

// Regression: without `dive`, per-item rules were never evaluated.
func TestOrderValidation_ValidatesItems(t *testing.T) {
	v := serviceutils.NewCustomValidator()
	base := Order{DestinationAddress: "somewhere"}

	bad := base
	bad.Items = []OrderProduct{{ProductID: validProductID, Quantity: -5}}
	if err := v.Validate(bad); err == nil {
		t.Error("negative quantity should be rejected")
	}

	missing := base
	missing.Items = []OrderProduct{{Quantity: 1}}
	if err := v.Validate(missing); err == nil {
		t.Error("missing product id should be rejected")
	}

	ok := base
	ok.Items = []OrderProduct{{ProductID: validProductID, Quantity: 2}}
	if err := v.Validate(ok); err != nil {
		t.Errorf("valid order rejected: %v", err)
	}
}

// Regression: product ids are UUID columns, so anything else used to reach
// Postgres and come back as a 500 instead of a 400.
func TestOrderValidation_RejectsNonUUIDProductID(t *testing.T) {
	v := serviceutils.NewCustomValidator()
	o := Order{DestinationAddress: "somewhere", Items: []OrderProduct{{ProductID: "p1", Quantity: 1}}}
	if err := v.Validate(o); err == nil {
		t.Error("non-UUID product id should be rejected")
	}
}

// Regression: orders.description is VARCHAR(255); longer input was a 500.
func TestOrderValidation_RejectsLongDescription(t *testing.T) {
	v := serviceutils.NewCustomValidator()
	o := Order{
		DestinationAddress: "somewhere",
		Description:        strings.Repeat("x", 256),
		Items:              []OrderProduct{{ProductID: validProductID, Quantity: 1}},
	}
	if err := v.Validate(o); err == nil {
		t.Error("description longer than 255 characters should be rejected")
	}
}
