package message_common

type MessageType string

// MessageInterface is implemented by every message payload struct.
// The message id and the wire format are handled outside of the payload.
type MessageInterface interface {
	GetMessageType() MessageType
}

type EncodeFunc func(id string, msg MessageInterface) ([]byte, error)

type DecodeFunc func(raw []byte) (string, MessageInterface, error)
