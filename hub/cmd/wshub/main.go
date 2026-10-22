package main

import (
	"fmt"

	"darbelis.eu/wshub/internal/di"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("TODO wshub")

	router := gin.Default()

	go di.GetWsHub().Run()
	// used just to swallow all incoming messages from WS clients
	go di.GetDelegatingMessageHandler().Handle(di.GetWsHub().GetIncomingMessagesChannel())

	_ = router.Run(":8585")
}
