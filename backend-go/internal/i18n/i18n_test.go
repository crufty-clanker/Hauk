package i18n

import (
	"testing"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		texts  Texts
		key    string
		expect string
	}{
		{English, "session_expired", "Session expired!"},
		{English, "incorrect_password", "Incorrect password!"},
		{nil, "session_invalid", "Invalid session!"}, // Falls back to English.
		{nil, "nonexistent_key", "nonexistent_key"},  // Returns key as last resort.
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := Translate(tt.texts, tt.key)
			if result != tt.expect {
				t.Fatalf("Translate(%v, %q) = %q, want %q", tt.texts, tt.key, result, tt.expect)
			}
		})
	}
}

func TestNegotiateLanguage(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"", "en"},
		{"en", "en"},
		{"en-US,en;q=0.9", "en"},           // "en-US" not in supported, falls back to "en"
		{"de-DE,de;q=0.9", "de"},            // "de-DE" not in supported, matches "de"
		{"nb_NO", "nb_NO"},                  // exact match
		{"unknown-lang", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NegotiateLanguage(tt.input)
			if result != tt.expect {
				t.Fatalf("NegotiateLanguage(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestShortLanguages(t *testing.T) {
	shorts := ShortLanguages()
	if len(shorts) != len(SupportedLanguages) {
		t.Fatalf("expected %d short languages, got %d", len(SupportedLanguages), len(shorts))
	}
	// Check that short languages don't contain underscores.
	for _, lang := range shorts {
		if len(lang) > 2 {
			t.Fatalf("short language should be 2 chars, got %q", lang)
		}
	}
}
