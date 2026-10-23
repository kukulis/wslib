package main

import (
	"darbelis.eu/wshub/internal/di"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	apiRoute := router.Group("/api")

	go di.GetWsHub().Run()
	// used just to swallow all incoming messages from WS clients
	go di.GetDelegatingMessageHandler().Handle(di.GetWsHub().GetIncomingMessagesChannel())

	di.RegisterWsRoutes(router)
	di.RegisterApiRoutes(apiRoute)

	_ = router.Run(":8585")
}
