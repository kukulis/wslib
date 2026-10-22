package helpers

import (
	"strings"
	"testing"
)

func TestValidateId(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		allowEmpty bool
		wantErr    bool
	}{
		{"valid alphanumeric", "abc123", false, false},
		{"valid with hyphen (uuid-like)", "550e8400-e29b-41d4-a716-446655440000", false, false},
		{"valid with underscore", "client_1", false, false},
		{"empty not allowed", "", false, true},
		{"empty allowed", "", true, false},
		{"exactly max length", strings.Repeat("a", maxIdLength), false, false},
		{"over max length", strings.Repeat("a", maxIdLength+1), false, true},
		{"contains space", "abc 123", false, true},
		{"contains slash", "abc/123", false, true},
		{"contains dot", "abc.123", false, true},
		{"contains unicode letter", "abc123é", false, true},
		{"contains newline", "abc\n123", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateId("field", tt.id, tt.allowEmpty)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateId(%q, allowEmpty=%v) expected error, got nil", tt.id, tt.allowEmpty)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateId(%q, allowEmpty=%v) unexpected error: %v", tt.id, tt.allowEmpty, err)
			}
		})
	}
}

// TestIdPatternCompiles guards against a broken regexp literal in idPattern.
// regexp.MustCompile panics at package init time on invalid syntax, which
// would surface only when something in this package actually runs (a
// build failure elsewhere wouldn't catch it) - so this test exercises the
// pattern directly to make that failure mode visible and attributable here.
func TestIdPatternCompiles(t *testing.T) {
	if idPattern == nil {
		t.Fatal("idPattern is nil")
	}
	if !idPattern.MatchString("abc123") {
		t.Error("expected idPattern to match a simple alphanumeric string")
	}
}
