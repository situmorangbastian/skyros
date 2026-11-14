package service

import "testing"

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  Alice@Example.COM "); got != "alice@example.com" {
		t.Fatalf("normalizeEmail = %q", got)
	}
}
