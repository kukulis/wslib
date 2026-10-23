package web

import (
	"darbelis.eu/wshub/internal/helpers"
	"darbelis.eu/wshub/pkg/ws"
	"github.com/gin-gonic/gin"
)

type WsController struct {
	hub *ws.Hub
}

func NewWsController(hub *ws.Hub) *WsController {
	return &WsController{hub: hub}
}

func (c *WsController) ServeWebSocketConnection(ctx *gin.Context) {
	clientId := ctx.Query("clientId")
	if err := helpers.ValidateId("clientId", clientId, false); err != nil {
		_ = ctx.AbortWithError(400, err)
		return
	}

	c.hub.RegisterWebSocketClient(clientId, ctx.Writer, ctx.Request)
}
