package web

import (
	"darbelis.eu/wshub/internal/helpers"
	"darbelis.eu/wshub/pkg/ws"
	"github.com/gin-gonic/gin"
)

type WsController struct {
	driversHub    *ws.Hub
	passengersHub *ws.Hub
}

func NewWsController(driversHub *ws.Hub, passengersHub *ws.Hub) *WsController {
	return &WsController{driversHub: driversHub, passengersHub: passengersHub}
}

func (c *WsController) ServeDriversWebSocketConnection(ctx *gin.Context) {
	clientId := ctx.Query("clientId")
	if err := helpers.ValidateId("clientId", clientId, false); err != nil {
		_ = ctx.AbortWithError(400, err)
		return
	}

	c.driversHub.RegisterWebSocketClient(clientId, ctx.Writer, ctx.Request)
}
