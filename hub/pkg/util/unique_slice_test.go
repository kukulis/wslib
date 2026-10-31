package util

import (
	"slices"
	"testing"
)

func TestUniqueSlice(t *testing.T) {
	if got := UniqueSlice([]string{"b", "a", "b", "c", "a"}); !slices.Equal(got, []string{"b", "a", "c"}) {
		t.Errorf("strings: got %v", got)
	}
	if got := UniqueSlice([]int{3, 3, 1, 2, 1}); !slices.Equal(got, []int{3, 1, 2}) {
		t.Errorf("ints: got %v", got)
	}
}

func TestUniqueSliceEmptyAndNil(t *testing.T) {
	if got := UniqueSlice([]string{}); got == nil || len(got) != 0 {
		t.Errorf("empty: got %#v", got)
	}
	if got := UniqueSlice[string](nil); got == nil || len(got) != 0 {
		t.Errorf("nil: got %#v", got)
	}
}

func TestUniqueSliceDoesNotChangeInput(t *testing.T) {
	input := []string{"a", "a", "b"}

	result := UniqueSlice(input)
	result[0] = "changed"

	if !slices.Equal(input, []string{"a", "a", "b"}) {
		t.Errorf("input was modified: %v", input)
	}
}
