package util

import "testing"

func TestDiffStruct(t *testing.T) {
	testCases := provideDiffStructTestCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			diff := DiffStruct(tc.a, tc.b)
			if diff != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, diff)
			}
		})
	}
}

type diffStructTestCase struct {
	name     string
	a        any
	b        any
	expected string
}

func provideDiffStructTestCases() []diffStructTestCase {
	return []diffStructTestCase{
		{
			name:     "empty",
			a:        struct{}{},
			b:        struct{}{},
			expected: "",
		},
		{
			name:     "two arrays",
			a:        []int{},
			b:        []int{},
			expected: "",
		},
	}
}
