package ws

// HubInterface is the subset of Hub's API used to send outgoing messages to
// clients and read incoming ones. It lets callers depend on this instead of
// the concrete *Hub, so tests can substitute HubMock.
type HubInterface interface {
	SendMessage(message ClientMessage)
	GetIncomingMessagesChannel() <-chan ClientMessage
}

var _ HubInterface = (*Hub)(nil)
