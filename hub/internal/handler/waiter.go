package handler

import (
	"darbelis.eu/wshub/pkg/message_common"
)

type Waiter struct {
	// TODO rename to connectionId
	clientId  string
	messageId string
	mChannel  chan message_common.MessageInterface
}
