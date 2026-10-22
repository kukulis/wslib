package util

import "testing"

func TestFilter1(t *testing.T) {
	data := []int{
		1, 5, 7, 10,
	}

	result := ArrayFilter(data, func(item int) bool { return item > 5 })

	expected := []int{
		7, 10,
	}

	equals := len(result) == len(expected)

	for i := 0; i < len(result); i++ {
		r := result[i]
		e := expected[i]

		equals = equals && r == e
	}

	if !equals {
		t.Error("Expected ", expected, " got ", result)
	}
}
