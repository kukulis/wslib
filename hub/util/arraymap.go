package util

func ArrayMap[T, U any](data []T, f func(T) U) []U {

	result := make([]U, 0)

	for _, t := range data {
		result = append(result, f(t))
	}

	return result
}
