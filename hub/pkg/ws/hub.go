// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"log/slog"
	"net/http"
	"sync"

	"darbelis.eu/wshub/pkg/message_common"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered connections.
	connections map[string]*ClientConnection

	// Register requests from the clients.di
	registerChannel chan *ClientConnection

	// Unregister requests from connections.
	unregisterChannel chan string

	incomingMessagesChannel chan ClientMessage
	outgoingMessagesChannel chan ClientMessage

	clientType string
	logger     *slog.Logger

	encodeFunc message_common.EncodeFunc
	decodeFunc message_common.DecodeFunc

	availableConnections     map[string]bool
	availableConnectionsLock sync.RWMutex
}

func NewHub(
	clientType string,
	encodeFunc message_common.EncodeFunc,
	decodeFunc message_common.DecodeFunc,

) *Hub {
	return &Hub{
		registerChannel:          make(chan *ClientConnection),
		unregisterChannel:        make(chan string, 256),
		connections:              make(map[string]*ClientConnection),
		incomingMessagesChannel:  make(chan ClientMessage, 256),
		outgoingMessagesChannel:  make(chan ClientMessage, 256),
		clientType:               clientType,
		logger:                   slog.Default().With("component", "Hubmessages.D", "client_type", clientType),
		encodeFunc:               encodeFunc,
		decodeFunc:               decodeFunc,
		availableConnections:     make(map[string]bool),
		availableConnectionsLock: sync.RWMutex{},
	}
}

// Run this makes registering, unregistering and message sending actions never happen in parallel
// need to find a way to get a list of clients from outside ... or do we?
// Can we mitigate a client's list in to an outside structure by using incoming messages from clients
// or adding hooks for various events?

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.registerChannel:
			h.connections[c.GetConnectionId()] = c

			h.addAvailableConnection(c.connectionId)

			// virtual message for client registration
			h.incomingMessagesChannel <- ClientMessage{
				ConnectionId: c.connectionId,
				Message: &message_common.ConnectionRegisteredMessage{
					ConnectionId: c.GetConnectionId(),
					ClientType:   h.clientType,
				},
			}
		case clientId := <-h.unregisterChannel:
			client, ok := h.connections[clientId]

			if !ok {
				h.logger.Warn("client not found to close", "client_id", clientId)
				break
			}
			delete(h.connections, clientId)

			h.removeAvailableConnection(clientId)

			// consider whether to use a wrapper Close function
			// TODO what happens if the client already closed?
			client.conn.Close()

			// sending virtual message to the hub using service
			h.incomingMessagesChannel <- ClientMessage{
				ConnectionId: client.GetConnectionId(),
				Message: &message_common.ConnectionUnregisteredMessage{
					ConnectionId: client.GetConnectionId(),
					ClientType:   h.clientType,
				},
			}

		case clientMessage := <-h.outgoingMessagesChannel:
			client, ok := h.connections[clientMessage.ConnectionId]
			if !ok {
				h.logger.Warn("client not found to send message to", "client_id", clientMessage.ConnectionId)
				break
			}

			sentOk := client.send(clientMessage)
			if !sentOk {
				delete(h.connections, clientMessage.ConnectionId)
				h.removeAvailableConnection(clientMessage.ConnectionId)
			}
		}
	}
}

// RegisterWebSocketConnection handles websocket requests from the peer.
func (h *Hub) RegisterWebSocketConnection(
	connectionId string,
	w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "client_id", connectionId, "error", err)
		return
	}

	c := NewClientConnection(connectionId, conn, h.incomingMessagesChannel, h.unregisterChannel, h.encodeFunc, h.decodeFunc)
	h.registerChannel <- c

	go c.writePump()
	go c.readPump()
}

func (h *Hub) SendMessage(message ClientMessage) {
	h.outgoingMessagesChannel <- message
}

// CloseConnection asks the hub to close the connection. The hub then handles it
// like any closed connection, including the ConnectionUnregisteredMessage.
func (h *Hub) CloseConnection(connectionId string) {
	h.unregisterChannel <- connectionId
}

func (h *Hub) GetIncomingMessagesChannel() <-chan ClientMessage {
	return h.incomingMessagesChannel
}

func (h *Hub) addAvailableConnection(
	connectionId string) {
	h.availableConnectionsLock.Lock()
	defer h.availableConnectionsLock.Unlock()

	h.availableConnections[connectionId] = true
}

func (h *Hub) removeAvailableConnection(
	connectionId string) {
	h.availableConnectionsLock.Lock()
	defer h.availableConnectionsLock.Unlock()

	delete(h.availableConnections, connectionId)
}

func (h *Hub) HasConnection(
	connectionId string) bool {
	h.availableConnectionsLock.RLock()
	defer h.availableConnectionsLock.RUnlock()

	return h.availableConnections[connectionId]
}
