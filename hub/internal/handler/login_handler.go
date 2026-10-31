package handler

import (
	"log/slog"

	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/pkg/message_common"
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
)

// LoginHandler handles the LoginMessage: it turns a pending connection into a
// logged in session. It is wired with DelegatingMessageHandler using SetHandler:
//
//	delegating.SetHandler(messages.MessageTypeLogin, loginHandler.HandleLogin)
//
// The connection is taken out of PendingConnections before the login is tried.
// Only one caller can take it, so the sweep and a login cannot both get the
// same connection: if the sweep was first, the login is refused, and if the
// login was first, the sweep no longer sees the connection.
type LoginHandler struct {
	pendingConnections *state.PendingConnections
	clientsRegistry    *state.ClientsRegistry
	hub                ws.HubInterface
	clock              util.Clock
	logger             *slog.Logger
}

func NewLoginHandler(
	pendingConnections *state.PendingConnections,
	clientsRegistry *state.ClientsRegistry,
	hub ws.HubInterface,
	clock util.Clock,
) *LoginHandler {
	return &LoginHandler{
		pendingConnections: pendingConnections,
		clientsRegistry:    clientsRegistry,
		hub:                hub,
		clock:              clock,
		logger:             slog.Default().With("handler", "LoginHandler"),
	}
}

// HandleLogin answers every LoginMessage with a LoginResultMessage that carries
// the id of the request. A failed login puts the connection back among the
// pending ones with its original connect time, so a client cannot keep itself
// open by failing again and again. A client that was logged in on another
// connection is moved to this one, and the old connection is closed.
func (h *LoginHandler) HandleLogin(connectionId string, messageId string, payload message_common.MessageInterface) {
	login, ok := payload.(*messages.LoginMessage)
	if !ok {
		h.logger.Error("login handler got a message of another type",
			"connection_id", connectionId, "message_type", payload.GetMessageType())
		return
	}

	pending, claimed := h.pendingConnections.Remove(connectionId)
	if !claimed {
		// already logged in, or swept out and being closed
		h.logger.Warn("login on a connection that is not pending", "connection_id", connectionId, "client_id", login.ClientId)
		h.reply(connectionId, messageId, false)
		return
	}

	session, replacedConnectionId, err := h.clientsRegistry.LoginSession(login.ClientId, login.Token, connectionId, h.clock.Now())
	if err != nil {
		h.logger.Warn("login failed", "connection_id", connectionId, "error", err)

		if addErr := h.pendingConnections.Add(connectionId, pending.RemoteIP, pending.ConnectedAt); addErr != nil {
			h.logger.Warn("closing a connection that cannot be put back", "connection_id", connectionId, "error", addErr)
			h.hub.CloseConnection(connectionId)
		}
		h.reply(connectionId, messageId, false)
		return
	}

	if replacedConnectionId != "" {
		h.logger.Info("client moved to a new connection",
			"client_id", session.ClientId, "connection_id", connectionId, "replaced_connection_id", replacedConnectionId)
		h.hub.CloseConnection(replacedConnectionId)
	}

	h.logger.Info("client logged in", "client_id", session.ClientId, "connection_id", connectionId)
	h.reply(connectionId, messageId, true)
}

func (h *LoginHandler) reply(connectionId string, messageId string, success bool) {
	h.hub.SendMessage(ws.ClientMessage{
		ConnectionId: connectionId,
		MessageId:    messageId,
		Message:      &messages.LoginResultMessage{Success: success},
	})
}
