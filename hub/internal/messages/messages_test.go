package messages

import (
	"encoding/json"
	"testing"
)

func TestHasMessageId_ReturnsStringValue(t *testing.T) {
	msg := NewDefaultMessage(MessageTypeDefault, "abc-123")

	if got := msg.GetMessageId(); got != "abc-123" {
		t.Errorf("expected %q, got %q", "abc-123", got)
	}
}

func TestHasMessageId_MissingKeyReturnsEmptyString(t *testing.T) {
	msg := DefaultMessage(make(map[string]interface{}))

	if got := msg.GetMessageId(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestHasMessageId_IntValueReturnsStringRepresentation(t *testing.T) {
	msg := DefaultMessage(make(map[string]interface{}))
	msg[MessageId] = 42

	if got := msg.GetMessageId(); got != "42" {
		t.Errorf("expected %q, got %q", "42", got)
	}
}

func TestHasMessageId_JSONNumericIdReturnsStringRepresentation(t *testing.T) {
	raw := []byte(`{"message_id":42}`)

	msg := DefaultMessage(make(map[string]interface{}))
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got := msg.GetMessageId(); got != "42" {
		t.Errorf("expected %q, got %q", "42", got)
	}
}

func TestHasMessageId_ArrayValueReturnsEmptyString(t *testing.T) {
	msg := DefaultMessage(make(map[string]interface{}))
	msg[MessageId] = []interface{}{"a", "b"}

	if got := msg.GetMessageId(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
