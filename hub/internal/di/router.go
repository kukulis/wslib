package di

import "github.com/gin-gonic/gin"

func RegisterWsRoutes(router *gin.Engine) {
	router.GET("/ws-old", func(c *gin.Context) { GetWsController().ServeWebSocketConnectionOld(c) })
}

func RegisterApiRoutes(apiRoute *gin.RouterGroup) {
	apiRoute.POST("/client-message", func(c *gin.Context) { GetApiController().SendClientMessage(c) })
}
