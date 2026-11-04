package main

import (
	"darbelis.eu/wshub/internal/di"
	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/pkg/message_common"
	"github.com/gin-gonic/gin"
)

func main() {
	di.GetDelegatingRegisteredMessageHandler().SetDefaultHandler(di.GetWaitResponseMessageHandler().Handle)

	di.GetDelegatingRegisteredMessageHandler().SetHandler(
		message_common.VirtualMessageTypeConnectionRegistered,
		di.GetConnectHandler().HandleConnectionRegistered,
	)
	di.GetDelegatingRegisteredMessageHandler().SetHandler(
		message_common.VirtualMessageTypeConnectionUnregistered,
		di.GetConnectHandler().HandleConnectionUnregistered,
	)
	
	di.GetDelegatingRegisteredMessageHandler().SetHandler(
		messages.MessageTypeLogin,
		di.GetLoginHandler().HandleLogin,
	)

	go di.GetDelegatingRegisteredMessageHandler().Handle(di.GetRegisteredWsHub().GetIncomingMessagesChannel())
	go di.GetRegisteredWsHub().Run()

	// the sweeps run as long as the process does, so the stop channel is never closed
	go di.GetSweepHandler().Run(make(chan struct{}))

	di.GetDelegatingMessageHandler().SetDefaultHandler(di.GetWaitResponseMessageHandler().Handle)
	go di.GetDelegatingMessageHandler().Handle(di.GetWsHub().GetIncomingMessagesChannel())
	go di.GetWsHub().Run()

	router := gin.Default()
	apiRoute := router.Group("/api")
	di.RegisterWsRoutes(router)
	di.RegisterApiRoutes(apiRoute)

	_ = router.Run(":8585")
}
