package message_common

type ClientRegisteredMessage struct {
	ClientId   string
	ClientType string
}

type ClientUnregisteredMessage struct {
	ClientId   string
	ClientType string
}

const (
	VirtualMessageTypeClientRegistered   MessageType = "client_registered"
	VirtualMessageTypeClientUnregistered MessageType = "client_unregistered"
)

func (ClientRegisteredMessage) GetMessageType() MessageType {
	return VirtualMessageTypeClientRegistered
}

func (ClientUnregisteredMessage) GetMessageType() MessageType {
	return VirtualMessageTypeClientUnregistered
}
