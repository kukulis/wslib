package handler

import (
	"fmt"
	"log/slog"

	"darbelis.eu/wshub/pkg/message_common"
	"darbelis.eu/wshub/pkg/ws"
)

type MessageHandlerFunction func(clientId string, payload message_common.MessageInterface)

type DelegatingMessageHandler struct {
	messageHandlers map[message_common.MessageType]MessageHandlerFunction
	logger          *slog.Logger
}

func NewDelegatingMessageHandler() *DelegatingMessageHandler {
	return &DelegatingMessageHandler{
		messageHandlers: make(map[message_common.MessageType]MessageHandlerFunction),
		logger:          slog.Default().With("handler", "DelegatingMessageHandler"),
	}
}

func (dhh *DelegatingMessageHandler) SetHandler(t message_common.MessageType, h MessageHandlerFunction) {
	dhh.messageHandlers[t] = h
}

func (dhh *DelegatingMessageHandler) Handle(messageChannel <-chan ws.ClientMessage) {
	for clientMessage := range messageChannel {
		t := clientMessage.Message.GetMessageType()
		handlerFunction, ok := dhh.messageHandlers[t]

		if !ok {
			dhh.logger.Warn("Ignored message type",
				"client_id", clientMessage.ClientId,
				"message_type", fmt.Sprintf("%T", clientMessage.Message),
			)
			continue
		}

		handlerFunction(clientMessage.ClientId, clientMessage.Message)
	}
}
