package di

import "github.com/gin-gonic/gin"

func RegisterWsRoutes(router *gin.Engine) {
	router.GET("/ws-old", func(c *gin.Context) { GetWsController().ServeWebSocketConnection(c) })

	// shouldn't it work?
	router.GET("/ws", GetWsControllerWithoutClient().ServeWebSocketConnection)
}

func RegisterApiRoutes(apiRoute *gin.RouterGroup) {
	//apiRoute.POST("/client-message", func(c *gin.Context) { GetApiController().SendClientMessage(c) })

	// shouldn't it work?
	apiRoute.POST("/client-message", GetApiController().SendClientMessage)

	apiRoute.POST("/registered-client-message", GetApiForWsController().SendClientMessage)
	apiRoute.POST("/tag-message", GetApiForWsController().SendTagMessage)
	apiRoute.POST("/register-session", GetApiForWsController().RegisterSession)
}
