package web

import (
	"encoding/json"
	"net/http"
	"time"

	"darbelis.eu/wshub/internal/handler"
	"darbelis.eu/wshub/internal/helpers"
	"darbelis.eu/wshub/internal/messages"
	ws2 "darbelis.eu/wshub/pkg/ws"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-gonic/gin"
)

type ApiController struct {
	hub                 *ws2.Hub
	waitResponseHandler *handler.WaitResponseMessageHandler
}

func NewApiController(hub *ws2.Hub, waitResponseHandler *handler.WaitResponseMessageHandler) *ApiController {
	return &ApiController{
		hub:                 hub,
		waitResponseHandler: waitResponseHandler,
	}
}

// TODO remake the whole function
func (c *ApiController) SendClientMessage(ctx *gin.Context) {
	clientId := ctx.Query("clientId")
	if err := helpers.ValidateId("clientId", clientId, false); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//log.Println("client id valid", clientId)

	// TODO duplicates with envelope.Id field
	messageId := ctx.Query("messageId")
	if err := helpers.ValidateId("messageId", messageId, true); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//log.Println("message id valid", messageId)

	// Parsing json allows sending valid json to the WS client too
	var envelope messages.Envelope
	if err := ctx.ShouldBindJSON(&envelope); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//log.Println("envelope valid")

	message := messages.NewDefaultMessage(string(envelope.Type), messageId)

	// payloads without fields may come without data
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, message); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	//log.Println("data valid")

	if !c.hub.HasConnection(clientId) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "there is no client with id " + clientId})
		return
	}
	//log.Println("client available")

	c.hub.SendMessage(ws2.ClientMessage{
		ConnectionId: clientId,
		MessageId:    messageId,
		Message:      message,
	})

	logger.Info("The wait for response value = ", envelope.WaitResponse, " messageId = ", messageId)
	// wait for response
	if envelope.WaitResponse > 0 {

		response := c.waitResponseHandler.WaitForResponse(clientId, messageId, time.Duration(int64(time.Millisecond)*int64(envelope.WaitResponse)))

		if response == nil {
			ctx.JSON(http.StatusGatewayTimeout, gin.H{"error": "no response from client within the given wait time"})
			return
		}

		ctx.JSON(http.StatusOK, response)
		return
	}

}
