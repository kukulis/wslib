package messages

import "darbelis.eu/wshub/message_common"

const (
	MessageTypeDefault = "default"
)

type DefaultMessage map[string]interface{}

func (*DefaultMessage) GetMessageType() message_common.MessageType {
	return MessageTypeDefault
}

func NewDefaultMessage() *DefaultMessage {
	msg := DefaultMessage(make(map[string]interface{}))

	return &msg
}
