package handler

import (
	"fmt"
	"log/slog"

	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/pkg/message_common"
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
)

// ConnectHandler handles the virtual messages the hub sends when a connection
// opens or closes: ConnectionRegisteredMessage and ConnectionUnregisteredMessage.
// It is wired with DelegatingMessageHandler using SetHandler:
//
//	delegating.SetHandler(message_common.VirtualMessageTypeConnectionRegistered, connectHandler.HandleConnectionRegistered)
//	delegating.SetHandler(message_common.VirtualMessageTypeConnectionUnregistered, connectHandler.HandleConnectionUnregistered)
type ConnectHandler struct {
	pendingConnections *state.PendingConnections
	clientsRegistry    *state.ClientsRegistry
	hub                ws.HubInterface
	clock              util.Clock
	logger             *slog.Logger
}

func NewConnectHandler(
	pendingConnections *state.PendingConnections,
	clientsRegistry *state.ClientsRegistry,
	hub ws.HubInterface,
	clock util.Clock,
) *ConnectHandler {
	return &ConnectHandler{
		pendingConnections: pendingConnections,
		clientsRegistry:    clientsRegistry,
		hub:                hub,
		clock:              clock,
		logger:             slog.Default().With("handler", "ConnectHandler"),
	}
}

// HandleConnectionRegistered puts the new connection among the pending ones; it
// stays there until it logs in, closes, or is swept out for waiting too long.
// A connection that cannot be tracked is closed, because nothing could ever
// time it out.
func (h *ConnectHandler) HandleConnectionRegistered(connectionId string, _ string, message message_common.MessageInterface) {
	// The hub sends a pointer. An empty RemoteIp is not counted against the limit
	// per address of PendingConnections.
	registeredMessage, ok := message.(*message_common.ConnectionRegisteredMessage)
	if !ok {
		// this is an Application misconfiguration case
		panic(fmt.Errorf("the registered message handle must receive ConnectionRegisteredInstance, %v", message))
	}
	if err := h.pendingConnections.Add(connectionId, registeredMessage.RemoteIp, h.clock.Now()); err != nil {
		h.logger.Warn("closing a connection that cannot be tracked", "connection_id", connectionId, "error", err)
		h.hub.CloseConnection(connectionId)
	}
}

// HandleConnectionUnregistered forgets the closed connection: it is removed from
// the pending ones if it never logged in, otherwise its session is detached from
// it. The session is kept, so the client can log in again on a new connection.
func (h *ConnectHandler) HandleConnectionUnregistered(connectionId string, _ string, _ message_common.MessageInterface) {
	h.pendingConnections.Remove(connectionId)

	if session, ok := h.clientsRegistry.DisconnectSessionByConnectionId(connectionId); ok {
		h.logger.Info("client disconnected", "connection_id", connectionId, "client_id", session.ClientId)
	}
}
