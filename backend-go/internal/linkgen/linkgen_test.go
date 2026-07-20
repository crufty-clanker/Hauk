package linkgen

import (
	"strings"
	"testing"
)

func TestGenerateUUIDv4(t *testing.T) {
	id := generateUUIDv4()
	if len(id) != 36 {
		t.Fatalf("expected UUID length 36, got %d", len(id))
	}
	// UUID format: 8-4-4-4-12
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Fatalf("expected 5 parts, got %d", len(parts))
	}
	if len(parts[0]) != 8 {
		t.Fatalf("first part should be 8 hex chars, got %d: %s", len(parts[0]), parts[0])
	}
	// Check version 4: the version nibble is set in byte 6.
	// Format: 8-4-4-4-12 → positions: 0-7, 9-12, 14-17, 19-22, 24-35.
	// Position 14 is the first hex char of the third group (b[6]).
	if id[14] != '4' {
		t.Fatalf("expected version 4 in UUID at pos 14, got %c", id[14])
	}
}

func TestGenerateHex(t *testing.T) {
	id8 := generateHex(8)
	if len(id8) != 16 {
		t.Fatalf("expected 16 hex chars for 8 bytes, got %d", len(id8))
	}
	for _, c := range id8 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("expected hex char, got %c", c)
		}
	}

	id16 := generateHex(16)
	if len(id16) != 32 {
		t.Fatalf("expected 32 hex chars for 16 bytes, got %d", len(id16))
	}
}

func TestGenerateAlphanumeric(t *testing.T) {
	// Test upper case no zero/O.
	id := generateAlphanumeric(16, upperCaseNoZeroO)
	if len(id) != 16 {
		t.Fatalf("expected length 16, got %d", len(id))
	}
	for _, c := range id {
		if c == '0' || c == 'O' {
			t.Fatalf("uppercase no-zero-O should not contain 0 or O, got %c", c)
		}
	}

	// Test lower case.
	id = generateAlphanumeric(16, lowerCase)
	if len(id) != 16 {
		t.Fatalf("expected length 16, got %d", len(id))
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			t.Fatalf("expected lower-case alphanumeric, got %c", c)
		}
	}

	// Test mixed case no I/L/O/0.
	id = generateAlphanumeric(32, mixedCaseNoILO0)
	if len(id) != 32 {
		t.Fatalf("expected length 32, got %d", len(id))
	}
	for _, c := range id {
		if c == 'I' || c == 'l' || c == 'O' || c == '0' {
			t.Fatalf("mixed-case should not contain I, l, O, or 0, got %c", c)
		}
	}
}

func TestGenerate4Plus4(t *testing.T) {
	id := generate4Plus4(upperCaseNoZeroO)
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts separated by dash, got %d", len(parts))
	}
	if len(parts[0]) != 4 || len(parts[1]) != 4 {
		t.Fatalf("expected 4-4 format, got %s", id)
	}
}

func TestValidateCustomLink(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"abc123", true},
		{"ABC-DEF", true},
		{"test_link", true},
		{"a-b-c", true},
		{"", false},
		{"has space", false},
		{"has!special", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ValidateCustomLink(tt.input); got != tt.valid {
				t.Fatalf("ValidateCustomLink(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestFormat4Plus4(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"ABCDEFGHIJ", "ABCD-EFGH"},
		{"AB", "AB"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Format4Plus4(tt.input); got != tt.expect {
				t.Fatalf("Format4Plus4(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}
