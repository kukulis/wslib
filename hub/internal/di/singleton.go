package di

import (
	"darbelis.eu/wshub/internal/handler"
	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/ws"
)

var wsHubInstance *ws.Hub = nil
var delegatingMessageHandler *handler.DelegatingMessageHandler = nil

func GetWsHub() *ws.Hub {
	if wsHubInstance == nil {
		wsHubInstance = ws.NewHub("common", messages.Encode, messages.Decode)
	}

	return wsHubInstance
}

func GetDelegatingMessageHandler() *handler.DelegatingMessageHandler {
	if delegatingMessageHandler == nil {
		delegatingMessageHandler = handler.NewDelegatingMessageHandler()
	}

	return delegatingMessageHandler
}
