// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"darbelis.eu/wshub/message_common"
	"darbelis.eu/wshub/util"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Register requests from the clients.
	registerChannel chan *Client

	// Unregister requests from clients.
	unregisterChannel chan string

	incomingMessagesChannel chan ClientMessage
	outgoingMessagesChannel chan ClientMessage

	// dispatcher not usable now but still may be required later
	dispatcher *util.Dispatcher
	clientType string
	logger     *slog.Logger

	encodeFunc message_common.EncodeFunc
	decodeFunc message_common.DecodeFunc
}

func NewHub(
	dispatcher *util.Dispatcher,
	clientType string,
	encodeFunc message_common.EncodeFunc,
	decodeFunc message_common.DecodeFunc,

) *Hub {
	return &Hub{
		registerChannel:         make(chan *Client),
		unregisterChannel:       make(chan string, 256),
		clients:                 make(map[string]*Client),
		incomingMessagesChannel: make(chan ClientMessage, 256),
		outgoingMessagesChannel: make(chan ClientMessage, 256),
		dispatcher:              dispatcher,
		clientType:              clientType,
		logger:                  slog.Default().With("component", "Hubmessages.D", "client_type", clientType),
		encodeFunc:              encodeFunc,
		decodeFunc:              decodeFunc,
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
			h.clients[c.GetClientId()] = c

			//h.dispatch(events.NewClientRegisteredEvent(c.GetClientId(), events.WithRegisteredClientType(h.clientType)))
			h.incomingMessagesChannel <- ClientMessage{
				ClientId: c.clientId,
				Message: &message_common.ClientRegisteredMessage{
					ClientId:   c.GetClientId(),
					ClientType: h.clientType,
				},
			}
		case clientId := <-h.unregisterChannel:
			client, ok := h.clients[clientId]

			if !ok {
				h.logger.Warn("client not found to close", "client_id", clientId)
				break
			}
			delete(h.clients, clientId)

			// consider whether to use a wrapper Close function
			client.conn.Close()

			h.incomingMessagesChannel <- ClientMessage{
				ClientId: client.GetClientId(),
				Message: &message_common.ClientUnregisteredMessage{
					ClientId:   client.GetClientId(),
					ClientType: h.clientType,
				},
			}
			//h.dispatch(events.NewClientUnregisteredEvent(client.GetClientId(), events.WithUnregisteredClientType(h.clientType)))

		case clientMessage := <-h.outgoingMessagesChannel:
			client, ok := h.clients[clientMessage.ClientId]
			if !ok {
				h.logger.Warn("client not found to send message to", "client_id", clientMessage.ClientId)
				break
			}

			sentOk := client.send(clientMessage)
			if !sentOk {
				delete(h.clients, clientMessage.ClientId)
			}
		}
	}
}

// dispatch recovers from a panic in a single listener so it can't kill the
// Hub's Run loop; only this one event is lost, not the whole hub.
func (h *Hub) dispatch(event util.Event) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("recovered from panic dispatching event",
				"event", event.GetName(),
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}()
	h.dispatcher.Dispatch(event)
}

// RegisterWebSocketClient handles websocket requests from the peer.
func (h *Hub) RegisterWebSocketClient(clientId string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "client_id", clientId, "error", err)
		return
	}

	c := NewClient(clientId, conn, h.incomingMessagesChannel, h.unregisterChannel, h.encodeFunc, h.decodeFunc)
	h.registerChannel <- c

	go c.writePump()
	go c.readPump()
}

func (h *Hub) SendMessage(message ClientMessage) {
	h.outgoingMessagesChannel <- message
}

func (h *Hub) GetIncomingMessagesChannel() <-chan ClientMessage {
	return h.incomingMessagesChannel
}
