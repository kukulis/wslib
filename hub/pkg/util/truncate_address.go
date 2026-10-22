package util

import "strings"

// TruncateAddress keeps only the first maxParts comma-separated segments of address,
// trimming whitespace around each. Nominatim addresses grow from specific to general
// (street, then neighbourhood, city, region, country, ...), so keeping a short prefix
// is enough for a human-readable address and drops the administrative tail.
// If address has maxParts segments or fewer, it is returned with segments trimmed but
// otherwise unchanged.
func TruncateAddress(address string, maxParts int) string {
	parts := strings.Split(address, ",")

	if maxParts < len(parts) {
		parts = parts[:maxParts]
	}

	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return strings.Join(parts, ", ")
}
