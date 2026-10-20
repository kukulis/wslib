package util

func ArrayFilter[T any](data []T, f func(T) bool) []T {
	result := make([]T, 0)

	for _, t := range data {

		if f(t) {
			result = append(result, t)
		}

	}
	return result
}
