package ws

import (
	"log/slog"
	"time"

	"darbelis.eu/wshub/message_common"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	clientId                   string
	outgoingThisClientMessages chan ClientMessage
	conn                       *websocket.Conn
	incomingMessages           chan ClientMessage
	// will pass client id when closed
	closedNotifier chan string
	logger         *slog.Logger
	encodeFunc     message_common.EncodeFunc
	decodeFunc     message_common.DecodeFunc
}

func (c *Client) GetClientId() string {
	return c.clientId
}

func NewClient(clientId string,
	conn *websocket.Conn,
	incomingMessages chan ClientMessage,
	closedNotifier chan string,
	encodeFunc message_common.EncodeFunc,
	decodeFunc message_common.DecodeFunc,
) *Client {
	return &Client{
		clientId:                   clientId,
		outgoingThisClientMessages: make(chan ClientMessage, 256),
		conn:                       conn,
		incomingMessages:           incomingMessages,
		closedNotifier:             closedNotifier,
		logger:                     slog.Default().With("component", "Client", "client_id", clientId),
		encodeFunc:                 encodeFunc,
		decodeFunc:                 decodeFunc,
	}
}

func (c *Client) send(msg ClientMessage) bool {
	select {
	case c.outgoingThisClientMessages <- msg:
		return true
	default:
		return false
	}
}

// ReadPump intended to run in a goroutine.
func (c *Client) readPump() {
	defer func() {
		c.closedNotifier <- c.clientId
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Warn("websocket read error", "error", err)
			}
			break
		}
		//messageBytes = bytes.TrimSpace(bytes.Replace(messageBytes, newline, space, -1))

		messageId, message, err := c.decodeFunc(messageBytes)
		if err != nil {
			// better handling later
			c.logger.Warn("message decode error", "error", err)
			continue
		}

		clientMessage := ClientMessage{
			ClientId:  c.clientId,
			Message:   message,
			MessageId: messageId,
		}

		c.incomingMessages <- clientMessage
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case clientMessage, ok := <-c.outgoingThisClientMessages:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			messageBytes, err := c.encodeFunc(clientMessage.MessageId, clientMessage.Message)

			if err != nil {
				c.logger.Error("message encode error", "error", err)
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.logger.Warn("websocket writer init error", "error", err)

				return
			}
			w.Write(messageBytes)

			if err := w.Close(); err != nil {
				c.logger.Warn("websocket writer close error", "error", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
