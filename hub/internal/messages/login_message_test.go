package messages

import (
	"encoding/json"
	"testing"
)

func TestDecode_LoginMessage(t *testing.T) {
	raw := []byte(`{"id":"m1","type":"login","data":{"clientId":"client-1","token":"secret"}}`)

	id, msg, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if id != "m1" {
		t.Errorf("expected id %q, got %q", "m1", id)
	}

	login, ok := msg.(*LoginMessage)
	if !ok {
		t.Fatalf("expected *LoginMessage, got %T", msg)
	}
	if login.ClientId != "client-1" || login.Token != "secret" {
		t.Errorf("unexpected payload: %+v", login)
	}
	if login.GetMessageType() != MessageTypeLogin {
		t.Errorf("unexpected type %q", login.GetMessageType())
	}
}

func TestDecode_LoginMessageWithoutDataIsEmpty(t *testing.T) {
	_, msg, err := Decode([]byte(`{"id":"m1","type":"login"}`))
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if login, ok := msg.(*LoginMessage); !ok || login.ClientId != "" || login.Token != "" {
		t.Errorf("expected an empty *LoginMessage, got %#v", msg)
	}
}

func TestEncode_LoginResultMessage(t *testing.T) {
	raw, err := Encode("m1", &LoginResultMessage{Success: true})
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	var envelope struct {
		Id   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Success bool `json:"success"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if envelope.Id != "m1" || envelope.Type != "login_result" || !envelope.Data.Success {
		t.Errorf("unexpected envelope: %s", raw)
	}
}
