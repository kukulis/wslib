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
type LoginResultMessage struct {
	Success bool `json:"success"`
}

func (LoginResultMessage) GetMessageType() message_common.MessageType {
	return MessageTypeLoginResult
}
