package truenas

import (
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		// Empty and zero
		{"", 0, false},
		{"0", 0, false},

		// Pure numeric (bytes)
		{"1024", 1024, false},
		{"1000000000", 1000000000, false},

		// SI units (go-humanize default)
		{"1KB", 1000, false},
		{"1MB", 1000000, false},
		{"1GB", 1000000000, false},
		{"10GB", 10000000000, false},
		{"500MB", 500000000, false},

		// Terabytes
		{"1TB", 1000000000000, false},
		{"2TB", 2000000000000, false},

		// Binary units (IEC)
		{"1KiB", 1024, false},
		{"1MiB", 1048576, false},
		{"1GiB", 1073741824, false},
		{"10GiB", 10737418240, false},
		{"1TiB", 1099511627776, false},

		// Case insensitive
		{"1gb", 1000000000, false},
		{"1gib", 1073741824, false},

		// With spaces
		{" 1GB ", 1000000000, false},

		// Invalid formats
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseSize(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}
