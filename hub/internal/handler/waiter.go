package handler

import (
	"darbelis.eu/wshub/pkg/message_common"
)

type Waiter struct {
	connectionId string
	messageId    string
	mChannel     chan message_common.MessageInterface
}
