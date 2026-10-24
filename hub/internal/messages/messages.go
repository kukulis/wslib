package messages

import (
	"strconv"

	"darbelis.eu/wshub/pkg/message_common"
)

const (
	MessageTypeDefault = "default"
	MessageTypeKey     = "message_type"
	MessageId          = "message_id"
)

type HasMessageId interface {
	message_common.MessageInterface
	GetMessageId() string
}

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

func (m *DefaultMessage) GetMessageId() string {
	switch messageId := (*m)[MessageId].(type) {
	case string:
		return messageId
	case int:
		return strconv.Itoa(messageId)
	case float64:
		// numeric message IDs decoded from JSON arrive as float64
		return strconv.FormatFloat(messageId, 'f', -1, 64)
	default:
		return ""
	}
}

func NewDefaultMessage(messageType string, messageId string) *DefaultMessage {
	msg := DefaultMessage(make(map[string]interface{}))
	msg[MessageTypeKey] = messageType
	msg[MessageId] = messageId

	return &msg
}
