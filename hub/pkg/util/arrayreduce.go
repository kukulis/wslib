package util

func ArrayReduce[T any](data []T, initialValue T, reductor func(a T, b T) T) T {
	result := initialValue

	for _, d := range data {
		result = reductor(result, d)
	}

	return result
}
