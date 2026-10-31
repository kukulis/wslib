package messages

import (
	"encoding/json"
	"fmt"

	"darbelis.eu/wshub/pkg/message_common"
)

func Encode(id string, msg message_common.MessageInterface) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return json.Marshal(Envelope{
		Id:   id,
		Type: msg.GetMessageType(),
		Data: data,
	})
}

// Decode returns messageId, the message, and the possible error
// The parameter 'raw' is supposed to be Envelope object
func Decode(raw []byte) (string, message_common.MessageInterface, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return "", nil, err
	}

	var msg message_common.MessageInterface

	switch env.Type {
	case MessageTypeLogin:
		msg = &LoginMessage{}
	default:
		msg = NewDefaultMessage(string(env.Type), env.Id)
	}

	// payloads without fields may come without data
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, msg); err != nil {
			return env.Id, nil, fmt.Errorf("decoding %q payload: %w", env.Type, err)
		}
	}

	return env.Id, msg, nil
}
