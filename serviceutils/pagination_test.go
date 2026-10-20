package serviceutils

import "testing"

func TestPageOffset(t *testing.T) {
	tests := []struct {
		name                 string
		page, size           int
		wantLimit, wantStart uint64
	}{
		{"first page", 1, 20, 20, 0},
		{"second page", 2, 20, 20, 20},
		{"zero page falls back to first", 0, 10, 10, 0},
		{"negative page falls back to first", -3, 10, 10, 0},
		{"zero size uses default", 3, 0, DefaultPageSize, 2 * DefaultPageSize},
		{"size is clamped", 2, 10_000, MaxPageSize, MaxPageSize},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset := PageOffset(tt.page, tt.size)
			if limit != tt.wantLimit || offset != tt.wantStart {
				t.Fatalf("PageOffset(%d, %d) = (%d, %d), want (%d, %d)",
					tt.page, tt.size, limit, offset, tt.wantLimit, tt.wantStart)
			}
		})
	}
}
