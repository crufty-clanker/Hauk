// Package linkgen generates share link IDs in various formats.
package linkgen

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/hauk/hauk-go/internal/store"
)

// Style represents the different link generation styles.
type Style int

const (
	Link4Plus4UpperCase   Style = 0
	Link4Plus4LowerCase   Style = 1
	Link4Plus4MixedCase   Style = 2
	LinkUUIDv4            Style = 3
	Link16Hex             Style = 4
	Link16UpperCase       Style = 5
	Link16LowerCase       Style = 6
	Link16MixedCase       Style = 7
	Link32Hex             Style = 8
	Link32UpperCase       Style = 9
	Link32LowerCase       Style = 10
	Link32MixedCase       Style = 11
)

// GenerateID generates a random link ID in the given style.
// It retries until a unique ID is found (not already in use in the store).
func GenerateID(style Style, st store.Store) (string, error) {
	for {
		id := generateLinkID(style)
		if _, found := st.Get(id); !found {
			return id, nil
		}
	}
}

func generateLinkID(style Style) string {
	switch style {
	case LinkUUIDv4:
		return generateUUIDv4()
	case Link16Hex:
		return generateHex(8)
	case Link16UpperCase:
		return generateAlphanumeric(16, upperCaseNoZeroO)
	case Link16LowerCase:
		return generateAlphanumeric(16, lowerCase)
	case Link16MixedCase:
		return generateAlphanumeric(16, mixedCaseNoILO0)
	case Link32Hex:
		return generateHex(16)
	case Link32UpperCase:
		return generateAlphanumeric(32, upperCaseNoZeroO)
	case Link32LowerCase:
		return generateAlphanumeric(32, lowerCase)
	case Link32MixedCase:
		return generateAlphanumeric(32, mixedCaseNoILO0)
	case Link4Plus4LowerCase:
		return generate4Plus4(lowerCase)
	case Link4Plus4MixedCase:
		return generate4Plus4(mixedCaseNoILO0)
	default: // Link4Plus4UpperCase (default)
		return generate4Plus4(upperCaseNoZeroO)
	}
}

// Character sets for link generation.
var (
	upperCaseNoZeroO = "123456789ABCDEFGHIJKLMNPQRSTUVWXYZ"
	lowerCase        = "0123456789abcdefghijklmnopqrstuvwxyz"
	mixedCaseNoILO0  = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz"
)

func generateUUIDv4() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to read random bytes: %v", err))
	}
	// Set version to 4.
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant to RFC4122.
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func generateHex(byteLen int) string {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to read random bytes: %v", err))
	}
	return fmt.Sprintf("%x", b)
}

func generateAlphanumeric(length int, charset string) string {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := range result {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			panic(fmt.Sprintf("failed to generate random number: %v", err))
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

func generate4Plus4(charset string) string {
	result := make([]byte, 8)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := range result {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			panic(fmt.Sprintf("failed to generate random number: %v", err))
		}
		result[i] = charset[n.Int64()]
	}
	s := string(result)
	return s[:4] + "-" + s[4:]
}

// ValidateCustomLink checks if a custom link ID is valid.
func ValidateCustomLink(id string) bool {
	if len(id) == 0 {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// IsReserved checks if a custom link ID is reserved for the given user.
func IsReserved(id string, reservedLinks map[string][]string, user string) bool {
	users, ok := reservedLinks[id]
	if !ok {
		return false
	}
	for _, u := range users {
		if u == user {
			return true
		}
	}
	return false
}

// Format4Plus4 formats a string as 4-4 with a dash separator.
func Format4Plus4(s string) string {
	if len(s) >= 8 {
		return s[:4] + "-" + s[4:8]
	}
	return s
}

// JoinPath joins a base URL with a path, ensuring proper trailing slash.
func JoinPath(base, path string) string {
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return base + path
}
