package messages

import (
	"encoding/json"
	"testing"
)

func TestDecode_RoundTrip(t *testing.T) {
	raw, err := Encode("msg-1", NewDefaultMessage())
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	id, msg, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if id != "msg-1" {
		t.Errorf("expected id %q, got %q", "msg-1", id)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}
	if msg.GetMessageType() != MessageTypeDefault {
		t.Errorf("expected type %q, got %q", MessageTypeDefault, msg.GetMessageType())
	}
}

func TestDecode_UnknownTypeFallsBackToDefaultMessage(t *testing.T) {
	raw := []byte(`{"id":"abc","type":"something-unknown","data":{"foo":"bar"}}`)

	id, msg, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if id != "abc" {
		t.Errorf("expected id %q, got %q", "abc", id)
	}

	dmp, ok := msg.(*DefaultMessage)
	if !ok {
		t.Fatalf("expected DefaultMessage, got %T", msg)
	}
	dm := *dmp
	if dm["foo"] != "bar" {
		t.Errorf("expected payload field foo=bar, got %v", dm["foo"])
	}
}

func TestDecode_EmptyDataIsAllowed(t *testing.T) {
	raw := []byte(`{"id":"no-data","type":"default"}`)

	id, msg, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if id != "no-data" {
		t.Errorf("expected id %q, got %q", "no-data", id)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}
}

func TestDecode_InvalidEnvelopeJSON(t *testing.T) {
	raw := []byte(`not-json`)

	_, _, err := Decode(raw)
	if err == nil {
		t.Fatal("expected error for invalid envelope JSON, got nil")
	}
}

func TestDecode_InvalidPayloadJSON(t *testing.T) {
	raw := []byte(`{"id":"bad-payload","type":"default","data":"not-an-object"}`)

	_, _, err := Decode(raw)
	if err == nil {
		t.Fatal("expected error for invalid payload JSON, got nil")
	}
}

func TestDecode_DataPreservedAsRawMessageOnEnvelope(t *testing.T) {
	// sanity check that Envelope.Data round-trips through json.RawMessage
	var env Envelope
	if err := json.Unmarshal([]byte(`{"id":"x","type":"default","data":{"a":1}}`), &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if len(env.Data) == 0 {
		t.Fatal("expected non-empty raw data")
	}
}
