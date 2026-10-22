package util

func ToMap[T any, K comparable](data []T, keyFunc func(T) K) map[K]T {
	result := make(map[K]T, len(data))

	for _, t := range data {
		result[keyFunc(t)] = t
	}

	return result
}
