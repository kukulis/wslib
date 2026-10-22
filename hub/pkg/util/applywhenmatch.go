package util

func ApplyWhenMatch[E, S any, K comparable](
	elements []E,
	sourceElements []S,
	destinationKeyGetter func(E) K,
	sourceKeyGetter func(S) K,
	action func(E, S),
	onMiss func(E),
) {
	indexedSources := ToMap(sourceElements, sourceKeyGetter)

	for _, element := range elements {
		destinationKey := destinationKeyGetter(element)

		source, ok := indexedSources[destinationKey]
		if !ok {
			onMiss(element)
			continue
		}

		action(element, source)
	}
}
