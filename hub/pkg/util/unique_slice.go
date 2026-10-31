package util

// UniqueSlice returns a new slice without repeated elements, keeping the order
// of the first occurrences.
func UniqueSlice[T comparable](items []T) []T {
	result := make([]T, 0, len(items))
	seen := make(map[T]struct{}, len(items))

	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		result = append(result, item)
	}

	return result
}
