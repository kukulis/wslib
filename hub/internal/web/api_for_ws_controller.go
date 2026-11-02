package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"darbelis.eu/wshub/internal/handler"
	"darbelis.eu/wshub/internal/helpers"
	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-gonic/gin"
)

type ApiForWsController struct {
	clientsRegistry     *state.ClientsRegistry
	clock               util.Clock
	hub                 *ws.Hub
	waitResponseHandler *handler.WaitResponseMessageHandler
}

func NewApiForWsController(
	clientsRegistry *state.ClientsRegistry,
	clock util.Clock,
	hub *ws.Hub,
	waitResponseHandler *handler.WaitResponseMessageHandler,
) *ApiForWsController {
	return &ApiForWsController{
		clientsRegistry:     clientsRegistry,
		clock:               clock,
		hub:                 hub,
		waitResponseHandler: waitResponseHandler,
	}
}

func (apiWs *ApiForWsController) RegisterSession(ctx *gin.Context) {

	clientId := ctx.Query("clientId")
	if err := helpers.ValidateId("clientId", clientId, false); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionToken := ctx.Query("sessionToken")
	if err := helpers.ValidateId("sessionToken", sessionToken, true); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tagsStr := ctx.Query("tags")

	tagsArray := strings.Split(tagsStr, ",")

	for _, tag := range tagsArray {
		err := helpers.ValidateId("tag", tag, false)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	session := state.ClientSession{
		ClientId: clientId,
		Token:    sessionToken,
		Tags:     tagsArray,
	}

	now := apiWs.clock.Now()

	err := apiWs.clientsRegistry.AddSession(session, now)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

func (apiWs *ApiForWsController) SendClientMessage(ctx *gin.Context) {
	clientId := ctx.Query("clientId")
	if err := helpers.ValidateId("clientId", clientId, false); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// XXX duplicates with envelope.Id field
	messageId := ctx.Query("messageId")
	if err := helpers.ValidateId("messageId", messageId, true); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// find connection id by a client id
	session, found := apiWs.clientsRegistry.FindSessionByClientId(clientId)

	if !found {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "The session was not found by clientId " + clientId})
		return
	}

	// Parsing json allows sending valid json to the WS client too
	var envelope messages.Envelope
	if err := ctx.ShouldBindJSON(&envelope); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//log.Println("envelope valid")

	// stage 2 of parsing message
	message := messages.NewDefaultMessage(string(envelope.Type), messageId)

	// payloads without fields may come without data
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, message); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	//log.Println("data valid")

	if !apiWs.hub.HasConnection(session.ConnectionId) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "there is no WS connection for client id " + clientId})
		return
	}
	//log.Println("client available")

	apiWs.hub.SendMessage(ws.ClientMessage{
		ConnectionId: session.ConnectionId,
		MessageId:    messageId,
		Message:      message,
	})

	logger.Info("The wait for response value = ", envelope.WaitResponse, " messageId = ", messageId)
	// wait for response ir required by the request
	if envelope.WaitResponse > 0 {

		response := apiWs.waitResponseHandler.WaitForResponse(session.ConnectionId, messageId, time.Duration(int64(time.Millisecond)*int64(envelope.WaitResponse)))

		if response == nil {
			ctx.JSON(http.StatusGatewayTimeout, gin.H{"error": "no response from client within the given wait time"})
			return
		}

		ctx.JSON(http.StatusOK, response)
		return
	}
}

// SendTagMessage sends the message to every logged in client carrying the tag,
// and answers with the number of connections it was handed to. A tag without
// logged in clients is not an error: the answer is {"delivered": 0}.
// The message cannot wait for a response, because one request would get many.
func (apiWs *ApiForWsController) SendTagMessage(ctx *gin.Context) {
	tag := ctx.Query("tag")
	if err := helpers.ValidateId("tag", tag, false); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messageId := ctx.Query("messageId")
	if err := helpers.ValidateId("messageId", messageId, true); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var envelope messages.Envelope
	if err := ctx.ShouldBindJSON(&envelope); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if envelope.WaitResponse > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "waitResponse is not supported for a message to a tag"})
		return
	}

	// the same message object goes to every connection; nobody changes it after this
	message := messages.NewDefaultMessage(string(envelope.Type), messageId)

	// payloads without fields may come without data
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, message); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	delivered := 0
	for _, session := range apiWs.clientsRegistry.FindSessionsByTag(tag) {
		// the session may outlive its connection for a moment
		if !apiWs.hub.HasConnection(session.ConnectionId) {
			continue
		}

		apiWs.hub.SendMessage(ws.ClientMessage{
			ConnectionId: session.ConnectionId,
			MessageId:    messageId,
			Message:      message,
		})
		delivered++
	}

	ctx.JSON(http.StatusOK, gin.H{"delivered": delivered})
}
