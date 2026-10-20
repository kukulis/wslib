package util

import (
	"fmt"
	"reflect"
	"strings"
)

// DiffStruct utility function, used to print failing assertions in tests
func DiffStruct(want, got any) string {
	rGot := reflect.ValueOf(got)
	rWant := reflect.ValueOf(want)
	if rGot.Kind() == reflect.Ptr {
		rGot = rGot.Elem()
	}
	if rWant.Kind() == reflect.Ptr {
		rWant = rWant.Elem()
	}
	rGotType := rGot.Type()
	var sb strings.Builder

	if rGotType.Kind() == reflect.Struct {
		for i := 0; i < rGot.NumField(); i++ {
			fieldGot := rGot.Field(i)
			fieldWant := rWant.Field(i)
			if !reflect.DeepEqual(fieldGot.Interface(), fieldWant.Interface()) {
				fmt.Fprintf(&sb, "  %-26s got %-30s want %s\n",
					rGotType.Field(i).Name+":",
					derefField(fieldGot),
					derefField(fieldWant),
				)
			}
		}
	} else if rGotType.Kind() == reflect.Slice {
		for i := 0; i < MaxInt(rGot.Len(), rWant.Len()); i++ {

			rGotValue := ""
			var rGotElem reflect.Value = reflect.Zero(rGotType.Elem())

			if i < rGot.Len() {
				rGotElem = rGot.Index(i)
				rGotValue = derefField(rGotElem)
			}

			rWantValue := ""
			var rWantElem reflect.Value = reflect.Zero(rGotType.Elem())

			if i < rWant.Len() {
				rWantElem = rWant.Index(i)
				rWantValue = derefField(rWantElem)
			}

			equalMarker := "<NOT EQUAL>"
			if reflect.DeepEqual(rGotElem.Interface(), rWantElem.Interface()) {
				equalMarker = "<EQUAL>"
			}
			fmt.Fprintf(&sb, "  %d:  %s |  want %s | got %s\n", i, equalMarker, rWantValue, rGotValue)

		}
	}
	return sb.String()
}

func derefField(v reflect.Value) string {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "nil"
		}
		return fmt.Sprintf("%v", v.Elem().Interface())
	}
	return fmt.Sprintf("%v", v.Interface())
}
