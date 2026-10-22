package ws

import (
	"darbelis.eu/wshub/pkg/message_common"
)

type ClientMessage struct {
	ClientId  string
	MessageId string
	Message   message_common.MessageInterface
}
