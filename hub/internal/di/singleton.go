package di

import (
	"darbelis.eu/wshub/internal/handler"
	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/internal/web"
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
)

var wsHubInstance *ws.Hub = nil
var delegatingMessageHandler *handler.DelegatingMessageHandler = nil
var wsController *web.WsController = nil
var apiController *web.ApiController = nil
var waitResponseMessageHandler *handler.WaitResponseMessageHandler = nil

var wsControllerWithoutClient *web.WsControllerWithoutClient = nil
var apiForWsController *web.ApiForWsController = nil
var registeredWsHub *ws.Hub = nil
var delegatingRegisteredMessageHandler *handler.DelegatingMessageHandler = nil

var clientsRegistry *state.ClientsRegistry = nil
var pendingConnections *state.PendingConnections = nil

var clock util.Clock = nil

//new ws controller new API controller
// separate hub , separate messageHandler , not sure if we need separate waitResponseMessageHandler or we may use same

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

func GetClientsRegistry() *state.ClientsRegistry {
	if clientsRegistry == nil {
		clientsRegistry = state.NewClientsRegistry()
	}
	return clientsRegistry
}

func GetPendingConnections() *state.PendingConnections {
	if pendingConnections == nil {
		pendingConnections = state.NewPendingConnections(10)
	}

	return pendingConnections
}

func GetApiForWsController() *web.ApiForWsController {

	if apiForWsController == nil {
		apiForWsController = web.NewApiForWsController(
			GetClientsRegistry(),
			GetClock(),
			GetRegisteredWsHub(),
			GetWaitResponseMessageHandler(),
		)
	}

	return apiForWsController
}

func GetClock() util.Clock {
	if clock == nil {
		clock = &util.RealClock{}
	}

	return clock
}

func GetRegisteredWsHub() *ws.Hub {
	if registeredWsHub == nil {
		registeredWsHub = ws.NewHub("registered", messages.Encode, messages.Decode)
	}

	return registeredWsHub
}

func GetDelegatingRegisteredMessageHandler() *handler.DelegatingMessageHandler {
	if delegatingRegisteredMessageHandler == nil {
		delegatingRegisteredMessageHandler = handler.NewDelegatingMessageHandler()
	}

	return delegatingRegisteredMessageHandler
}

func GetWsControllerWithoutClient() *web.WsControllerWithoutClient {
	if wsControllerWithoutClient == nil {
		wsControllerWithoutClient = web.NewWsControllerWithoutClient(GetRegisteredWsHub())
	}

	return wsControllerWithoutClient
}
