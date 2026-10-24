package messages

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestDecode_RoundTrip(t *testing.T) {
	raw, err := Encode("msg-1", NewDefaultMessage(MessageTypeDefault, ""))
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

func TestDecode_NestedJSONPayload(t *testing.T) {
	raw := []byte(`{"id":"nested-1","type":"something-unknown","data":{"foo":{"bar":"baz"},"list":[1,2,3]}}`)

	id, msg, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if id != "nested-1" {
		t.Errorf("expected id %q, got %q", "nested-1", id)
	}

	dmp, ok := msg.(*DefaultMessage)
	if !ok {
		t.Fatalf("expected *DefaultMessage, got %T", msg)
	}
	dm := *dmp

	foo, ok := dm["foo"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected foo to be an object, got %T", dm["foo"])
	}
	if foo["bar"] != "baz" {
		t.Errorf("expected foo.bar=baz, got %v", foo["bar"])
	}

	list, ok := dm["list"].([]interface{})
	if !ok {
		t.Fatalf("expected list to be an array, got %T", dm["list"])
	}
	if len(list) != 3 {
		t.Errorf("expected list of length 3, got %d", len(list))
	}
}

func TestEncodeDecode_RoundTripNoInformationLoss(t *testing.T) {
	msg := NewDefaultMessage("custom-type", "")
	(*msg)["foo"] = "bar"
	(*msg)["nested"] = map[string]interface{}{"a": float64(1)}
	(*msg)["list"] = []interface{}{float64(1), float64(2), float64(3)}

	raw, err := Encode("round-trip-id", msg)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	id, decoded, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if id != "round-trip-id" {
		t.Errorf("expected id %q, got %q", "round-trip-id", id)
	}

	dmp, ok := decoded.(*DefaultMessage)
	if !ok {
		t.Fatalf("expected *DefaultMessage, got %T", decoded)
	}

	dm := *dmp

	if dm["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", dm["foo"])
	}
	nested, ok := dm["nested"].(map[string]interface{})
	if !ok || nested["a"] != float64(1) {
		t.Errorf("expected nested.a=1, got %v", dm["nested"])
	}
	list, ok := dm["list"].([]interface{})
	if !ok || len(list) != 3 {
		t.Errorf("expected list of length 3, got %v", dm["list"])
	}
	if dm[MessageTypeKey] != "custom-type" {
		t.Errorf("expected %s=%q, got %v", MessageTypeKey, "custom-type", dm[MessageTypeKey])
	}

	// Re-encoding the decoded message must reproduce the exact same bytes:
	// nothing was lost, and the only "extra" field is MessageTypeKey, which
	// is expected to be present in the payload both times.
	raw2, err := Encode(id, decoded)
	if err != nil {
		t.Fatalf("second Encode returned error: %v", err)
	}
	if !bytes.Equal(raw, raw2) {
		t.Errorf("round-trip bytes differ:\n first: %s\nsecond: %s", raw, raw2)
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
