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

func Decode(raw []byte) (string, message_common.MessageInterface, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return "", nil, err
	}

	var msg message_common.MessageInterface

	switch env.Type {
	default:
		msg = NewDefaultMessage(string(env.Type))
	}

	// payloads without fields may come without data
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, msg); err != nil {
			return env.Id, nil, fmt.Errorf("decoding %q payload: %w", env.Type, err)
		}
	}

	return env.Id, msg, nil
}
