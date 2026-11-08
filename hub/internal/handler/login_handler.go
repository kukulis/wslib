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
//
// Every successful login replaces the token of the session with a new one, which
// goes to the client in the LoginResultMessage. The token in the page source
// works once, so a third party that copied it cannot log in after the real
// client did. A session that is bound to a live connection cannot be logged in
// on another one, so the real client cannot be pushed off either.
type LoginHandler struct {
	pendingConnections *state.PendingConnections
	clientsRegistry    *state.ClientsRegistry
	hub                ws.HubInterface
	clock              util.Clock
	tokenGenerator     util.IdGenerator
	logger             *slog.Logger
}

func NewLoginHandler(
	pendingConnections *state.PendingConnections,
	clientsRegistry *state.ClientsRegistry,
	hub ws.HubInterface,
	clock util.Clock,
	tokenGenerator util.IdGenerator,
) *LoginHandler {
	return &LoginHandler{
		pendingConnections: pendingConnections,
		clientsRegistry:    clientsRegistry,
		hub:                hub,
		clock:              clock,
		tokenGenerator:     tokenGenerator,
		logger:             slog.Default().With("handler", "LoginHandler"),
	}
}

// HandleLogin answers every LoginMessage with a LoginResultMessage that carries
// the id of the request. A failed login puts the connection back among the
// pending ones with its original connect time, so a client cannot keep itself
// open by failing again and again. A login on a session that already has a live
// connection fails too: the old connection stays, and the session is logged in
// again only after the old connection is gone. A successful login carries the
// new token of the session.
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
		h.reply(connectionId, messageId, false, "")
		return
	}

	session, err := h.clientsRegistry.LoginSession(login.ClientId, login.Token, connectionId, h.clock.Now())
	if err != nil {
		h.logger.Warn("login failed", "connection_id", connectionId, "error", err)
		h.refuse(pending, connectionId, messageId)
		return
	}

	// the token of this login stops working here; the reply brings the next one
	newToken := h.tokenGenerator.NextId()
	rotated := h.clientsRegistry.UpdateSessionByConnectionId(connectionId, func(session *state.ClientSession) {
		session.Token = newToken
	})
	if !rotated {
		// the session was removed between the login and the rotation
		h.logger.Warn("session vanished during login", "connection_id", connectionId, "client_id", session.ClientId)
		h.refuse(pending, connectionId, messageId)
		return
	}

	h.logger.Info("client logged in", "client_id", session.ClientId, "connection_id", connectionId)
	h.reply(connectionId, messageId, true, newToken)
}

// refuse answers a failed login and puts the connection back among the pending ones.
func (h *LoginHandler) refuse(pending state.PendingConnection, connectionId string, messageId string) {
	if addErr := h.pendingConnections.Add(connectionId, pending.RemoteIP, pending.ConnectedAt); addErr != nil {
		h.logger.Warn("closing a connection that cannot be put back", "connection_id", connectionId, "error", addErr)
		h.hub.CloseConnection(connectionId)
	}
	h.reply(connectionId, messageId, false, "")
}

func (h *LoginHandler) reply(connectionId string, messageId string, success bool, token string) {
	h.hub.SendMessage(ws.ClientMessage{
		ConnectionId: connectionId,
		MessageId:    messageId,
		Message:      &messages.LoginResultMessage{Success: success, Token: token},
	})
}
