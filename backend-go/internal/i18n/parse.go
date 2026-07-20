package i18n

import (
	"strconv"
	"strings"
)

// parseFloat parses a float64 from a string.
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	return strconv.ParseFloat(s, 64)
}
