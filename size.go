package truenas

import (
	"strings"

	"github.com/dustin/go-humanize"
)

// ParseSize parses a size string to bytes using go-humanize.
// Accepts formats like "10GB", "500MB", "1TB", "1024", "10GiB", etc.
// See https://pkg.go.dev/github.com/dustin/go-humanize#ParseBytes
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0, nil
	}

	bytes, err := humanize.ParseBytes(s)
	if err != nil {
		return 0, err
	}

	return int64(bytes), nil
}
