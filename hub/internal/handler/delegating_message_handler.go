package handler

import (
	"fmt"
	"log/slog"

	"darbelis.eu/wshub/pkg/message_common"
	"darbelis.eu/wshub/pkg/ws"
)

type MessageHandlerFunction func(clientId string, messageId string, payload message_common.MessageInterface)

type DelegatingMessageHandler struct {
	messageHandlers map[message_common.MessageType]MessageHandlerFunction
	logger          *slog.Logger
	defaultHandler  MessageHandlerFunction
}

func NewDelegatingMessageHandler() *DelegatingMessageHandler {
	dhh := &DelegatingMessageHandler{
		messageHandlers: make(map[message_common.MessageType]MessageHandlerFunction),
		logger:          slog.Default().With("handler", "DelegatingMessageHandler"),
	}

	dhh.defaultHandler = dhh.defaultMessageHandlerFunction

	return dhh
}

func (dhh *DelegatingMessageHandler) defaultMessageHandlerFunction(clientId string, messageId string, payload message_common.MessageInterface) {
	dhh.logger.Warn("Ignored message type from client",
		"client_id:", clientId,
		"message_id", messageId,
		"message_type", fmt.Sprintf("%s", payload.GetMessageType()))
}

func (dhh *DelegatingMessageHandler) SetHandler(t message_common.MessageType, h MessageHandlerFunction) {
	dhh.messageHandlers[t] = h
}

func (dhh *DelegatingMessageHandler) SetDefaultHandler(h MessageHandlerFunction) {
	dhh.defaultHandler = h
}

func (dhh *DelegatingMessageHandler) Handle(messageChannel <-chan ws.ClientMessage) {
	for clientMessage := range messageChannel {
		t := clientMessage.Message.GetMessageType()
		handlerFunction, ok := dhh.messageHandlers[t]

		if !ok {
			dhh.defaultHandler(clientMessage.ClientId, clientMessage.MessageId, clientMessage.Message)
			continue
		}
		handlerFunction(clientMessage.ClientId, clientMessage.MessageId, clientMessage.Message)
	}
}
