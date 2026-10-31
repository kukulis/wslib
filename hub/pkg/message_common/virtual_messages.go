package message_common

type ConnectionRegisteredMessage struct {
	ConnectionId string
	ClientType   string
}

type ConnectionUnregisteredMessage struct {
	ConnectionId string
	ClientType   string
}

const (
	VirtualMessageTypeConnectionRegistered   MessageType = "connection_registered"
	VirtualMessageTypeConnectionUnregistered MessageType = "connection_unregistered"
)

func (ConnectionRegisteredMessage) GetMessageType() MessageType {
	return VirtualMessageTypeConnectionRegistered
}

func (ConnectionUnregisteredMessage) GetMessageType() MessageType {
	return VirtualMessageTypeConnectionUnregistered
}
