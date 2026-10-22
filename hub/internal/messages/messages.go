package messages

import (
	"darbelis.eu/wshub/pkg/message_common"
)

const (
	MessageTypeDefault = "default"
	MessageTypeKey     = "message_type"
)

type DefaultMessage map[string]interface{}

func (m *DefaultMessage) GetMessageType() message_common.MessageType {
	injectedMessageType, ok := (*m)[MessageTypeKey]

	if !ok {
		return MessageTypeDefault
	}

	strMessageType, ok := injectedMessageType.(string)

	if !ok {
		return MessageTypeDefault
	}

	return message_common.MessageType(strMessageType)
}

func NewDefaultMessage(messageType string) *DefaultMessage {
	msg := DefaultMessage(make(map[string]interface{}))
	msg[MessageTypeKey] = messageType

	return &msg
}
