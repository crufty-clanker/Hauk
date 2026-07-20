// Package i18n handles internationalization for the Hauk backend.
package i18n

import (
	"strings"
)

// Texts holds all translated strings for a language.
type Texts map[string]string

// SupportedLanguages lists the supported language codes.
var SupportedLanguages = []string{"ca", "de", "en", "eu", "fr", "it", "nb_NO", "nl", "nn", "ro", "ru", "tr", "uk"}

// ShortLanguages returns just the language codes (without country).
func ShortLanguages() []string {
	result := make([]string, 0, len(SupportedLanguages))
	for _, lang := range SupportedLanguages {
		if idx := strings.Index(lang, "_"); idx >= 0 {
			result = append(result, lang[:idx])
		} else {
			result = append(result, lang)
		}
	}
	return result
}

// English contains the English translations (used as fallback).
var English = Texts{
	"config_missing":            "Unable to find config!",
	"session_expired":           "Session expired!",
	"share_not_found":           "The given share does not exist!",
	"group_share_not_adoptable": "You cannot adopt group shares!",
	"share_adoption_not_allowed": "The host of the given share does not permit adoption!",
	"incorrect_password":        "Incorrect password!",
	"username_required":         "Username required!",
	"share_too_long":            "Share period is too long!",
	"interval_too_long":         "Update interval is too long!",
	"interval_too_short":        "Update interval is too short!",
	"share_mode_unsupported":    "Unsupported share mode!",
	"group_pin_invalid":         "Invalid group PIN!",
	"session_invalid":           "Invalid session!",
	"location_invalid":          "Invalid location!",
	"group_e2e_unsupported":     "Group shares cannot be password protected!",
	"e2e_adoption_not_allowed":  "This share is password protected and cannot be adopted!",
	"ldap_extension_missing":    "The LDAP extension is not available!",
	"ldap_config_error":         "Failed to set LDAP connection parameters!",
	"ldap_connection_failed":    "Failed to connect to the LDAP server!",
	"ldap_search_failed":        "Failed to look up user on the LDAP server!",
	"ldap_user_unauthorized":    "User not found, not authorized, or incorrect password!",
	"ldap_search_ambiguous":     "Matched multiple users - the LDAP filter is too broad!",
	"cannot_find_password_file": "Cannot find password file!",
}

// Translate returns the translated text for the given key, falling back to English.
func Translate(texts Texts, key string) string {
	if v, ok := texts[key]; ok {
		return v
	}
	if v, ok := English[key]; ok {
		return v
	}
	return key // Return key itself as last resort.
}

// NegotiateLanguage extracts the best matching language from the Accept-Language header.
func NegotiateLanguage(acceptLang string) string {
	if acceptLang == "" {
		return "en"
	}

	// Parse Accept-Language header.
	// Format: "en-US,en;q=0.9,de;q=0.8"
	type langEntry struct {
		full string
		short string
		q    float64
	}

	var entries []langEntry
	parts := strings.Split(acceptLang, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		subparts := strings.Split(part, ";")
		full := strings.TrimSpace(subparts[0])
		q := 1.0

		for _, sp := range subparts[1:] {
			sp = strings.TrimSpace(sp)
			if strings.HasPrefix(sp, "q=") {
				if v, err := parseFloat(sp[2:]); err == nil {
					q = v
				}
			}
		}

		// Extract short language code.
		short := full
		if idx := strings.Index(full, "-"); idx >= 0 {
			short = full[:idx]
		}
		if idx := strings.Index(full, "_"); idx >= 0 {
			short = full[:idx]
		}

		entries = append(entries, langEntry{full: full, short: short, q: q})
	}

	// Find the best matching language.
	best := "en"
	bestQ := 0.0

	for _, entry := range entries {
		// Check if full language code matches exactly.
		if matches(entry.full, SupportedLanguages) && entry.q > bestQ {
			best = entry.full
			bestQ = entry.q
			continue
		}
		// Check if short language code matches any supported language
		// (either exact match or prefix with underscore, e.g. "nb" matches "nb_NO").
		if matches(entry.short, ShortLanguages()) && entry.q*0.9 > bestQ {
			for _, sl := range SupportedLanguages {
				if sl == entry.short || strings.HasPrefix(sl, entry.short+"_") {
					best = sl
					bestQ = entry.q * 0.9
					break
				}
			}
		}
	}

	return best
}

func matches(s string, list []string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}


