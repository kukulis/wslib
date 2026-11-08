package messages

import "darbelis.eu/wshub/pkg/message_common"

const (
	MessageTypeLogin       message_common.MessageType = "login"
	MessageTypeLoginResult message_common.MessageType = "login_result"
)

// LoginMessage is the first message of a client: it proves the connection
// belongs to the provisioned client id, with the token the backend gave to it.
type LoginMessage struct {
	ClientId string `json:"clientId"`
	Token    string `json:"token"`
}

func (LoginMessage) GetMessageType() message_common.MessageType {
	return MessageTypeLogin
}

// LoginResultMessage answers a LoginMessage. It carries no reason on purpose:
// the reason of a failure goes to the hub's log only, so a client cannot use it
// to find out which client ids exist.
//
// A successful login also carries the token for the next login of the client,
// for example after a reconnect. The token it just logged in with does not work
// any more, so a copy of it, taken from the page source, is worthless. A failed
// login carries no token.
type LoginResultMessage struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
}

func (LoginResultMessage) GetMessageType() message_common.MessageType {
	return MessageTypeLoginResult
}
