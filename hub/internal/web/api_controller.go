package web

import (
	"encoding/json"
	"net/http"

	"darbelis.eu/wshub/internal/helpers"
	"darbelis.eu/wshub/internal/messages"
	ws2 "darbelis.eu/wshub/pkg/ws"
	"github.com/gin-gonic/gin"
)

type ApiController struct {
	hub *ws2.Hub
}

func NewApiController(hub *ws2.Hub) *ApiController {
	return &ApiController{
		hub: hub,
	}
}

func (c *ApiController) SendClientMessage(ctx *gin.Context) {
	clientId := ctx.Query("clientId")
	if err := helpers.ValidateId("clientId", clientId, false); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//log.Println("client id valid", clientId)

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

	message := messages.NewDefaultMessage(string(envelope.Type))

	// payloads without fields may come without data
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, message); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	//log.Println("data valid")

	if !c.hub.HasClient(clientId) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "there is no client with id " + clientId})
		return
	}
	//log.Println("client available")

	c.hub.SendMessage(ws2.ClientMessage{
		ClientId:  clientId,
		MessageId: messageId,
		Message:   message,
	})

}
