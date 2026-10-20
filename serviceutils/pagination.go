package serviceutils

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// PageOffset converts a 1-based page and a page size into a normalised
// (limit, offset) pair. Non-positive pages fall back to the first page and the
// page size is clamped to [1, MaxPageSize].
func PageOffset(page, pageSize int) (limit, offset uint64) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	if page < 1 {
		page = 1
	}
	return uint64(pageSize), uint64((page - 1) * pageSize)
}
