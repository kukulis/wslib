package main

import (
	"darbelis.eu/wshub/internal/di"
	"github.com/gin-gonic/gin"
)

func main() {

	di.GetDelegatingMessageHandler().SetDefaultHandler(di.GetWaitResponseMessageHandler().Handle)
	go di.GetDelegatingMessageHandler().Handle(di.GetWsHub().GetIncomingMessagesChannel())
	go di.GetWsHub().Run()

	router := gin.Default()
	apiRoute := router.Group("/api")
	di.RegisterWsRoutes(router)
	di.RegisterApiRoutes(apiRoute)

	_ = router.Run(":8585")
}
