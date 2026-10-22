package helpers

import (
	"fmt"
	"regexp"
)

const maxIdLength = 64

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidateId checks that id is alphanumeric (allowing '-' and '_', to
// accommodate UUID-formatted ids), within maxIdLength, and non-empty unless
// allowEmpty is true.
func ValidateId(fieldName, id string, allowEmpty bool) error {
	if id == "" {
		if allowEmpty {
			return nil
		}
		return fmt.Errorf("%s must not be empty", fieldName)
	}

	if len(id) > maxIdLength {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, maxIdLength)
	}

	if !idPattern.MatchString(id) {
		return fmt.Errorf("%s must be alphanumeric (may include '-' and '_')", fieldName)
	}

	return nil
}
