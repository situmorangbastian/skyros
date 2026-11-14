package models

import (
	"testing"

	"github.com/situmorangbastian/skyros/serviceutils"
)

// Regression: `required` alone only rejects zero, so negative prices were stored.
func TestProductValidation_Price(t *testing.T) {
	v := serviceutils.NewCustomValidator()
	for _, tt := range []struct {
		price   int64
		wantErr bool
	}{
		{price: -100, wantErr: true},
		{price: 0, wantErr: true},
		{price: 1, wantErr: false},
	} {
		err := v.Validate(Product{Name: "n", Description: "d", Price: tt.price})
		if (err != nil) != tt.wantErr {
			t.Errorf("price %d: err = %v, wantErr %v", tt.price, err, tt.wantErr)
		}
	}
}
