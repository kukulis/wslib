package test_utils

import (
	"reflect"
	"testing"
)

func AssertEquals(expected, received interface{}, t *testing.T, fatal bool, message string) {
	if !reflect.DeepEqual(expected, received) {
		if fatal {
			t.Fatalf(message+" Expected %v, received %v", expected, received)
		} else {
			t.Errorf(message+" Expected %v, received %v", expected, received)
		}
	}
}
