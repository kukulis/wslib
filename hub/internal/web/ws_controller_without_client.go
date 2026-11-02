package web

import (
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
	"github.com/gin-gonic/gin"
)

type WsControllerWithoutClient struct {
	hub *ws.Hub
}

func NewWsControllerWithoutClient(hub *ws.Hub) *WsControllerWithoutClient {
	return &WsControllerWithoutClient{hub: hub}
}

func (c *WsControllerWithoutClient) ServeWebSocketConnection(ctx *gin.Context) {

	connectionId, err := util.RandomHash(16)

	if err != nil {
		_ = ctx.AbortWithError(500, err)
		return
	}

	c.hub.RegisterWebSocketConnection(connectionId, ctx.Writer, ctx.Request)
}
