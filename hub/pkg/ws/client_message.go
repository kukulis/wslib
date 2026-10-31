package ws

import (
	"darbelis.eu/wshub/pkg/message_common"
)

type ClientMessage struct {
	ConnectionId string
	MessageId    string
	Message      message_common.MessageInterface
}
