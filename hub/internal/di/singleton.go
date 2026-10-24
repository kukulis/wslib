package di

import (
	"darbelis.eu/wshub/internal/handler"
	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/internal/web"
	"darbelis.eu/wshub/pkg/ws"
)

var wsHubInstance *ws.Hub = nil
var delegatingMessageHandler *handler.DelegatingMessageHandler = nil
var wsController *web.WsController = nil
var apiController *web.ApiController = nil
var waitResponseMessageHandler *handler.WaitResponseMessageHandler = nil

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

func GetWsController() *web.WsController {
	if wsController == nil {
		wsController = web.NewWsController(GetWsHub())
	}

	return wsController
}

func GetApiController() *web.ApiController {
	if apiController == nil {
		apiController = web.NewApiController(GetWsHub(), GetWaitResponseMessageHandler())
	}
	return apiController
}

func GetWaitResponseMessageHandler() *handler.WaitResponseMessageHandler {
	if waitResponseMessageHandler == nil {
		waitResponseMessageHandler = handler.NewWaitResponseMessageHandler()
	}

	return waitResponseMessageHandler
}
