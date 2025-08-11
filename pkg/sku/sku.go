package sku

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

func GenerateSKUCode(name string, prefix ...string) string {
	// Normalize: uppercase, remove non-alphanum, collapse spaces
	re := regexp.MustCompile(`[^A-Za-z0-9]+`)
	normalized := re.ReplaceAllString(strings.ToUpper(name), "")

	// If too long, hash the name to ensure uniqueness and indexability
	const maxLen = 32
	if len(normalized) > maxLen {
		h := sha1.New()
		h.Write([]byte(normalized))
		normalized = hex.EncodeToString(h.Sum(nil))[:maxLen]
	}

	// Add prefix if provided
	if len(prefix) > 0 && prefix[0] != "" {
		return fmt.Sprintf("%s_%s", strings.ToUpper(prefix[0]), normalized)
	}
	return normalized
}
