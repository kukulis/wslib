package di

import "github.com/gin-gonic/gin"

func RegisterWsRoutes(router *gin.Engine) {
	router.GET("/ws", func(c *gin.Context) { GetWsController().ServeWebSocketConnection(c) })
}

func RegisterApiRoutes(apiRoute *gin.RouterGroup) {
	apiRoute.POST("/client-message", func(c *gin.Context) { GetApiController().SendClientMessage(c) })
}
