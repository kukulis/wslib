package ws

// HubMock implements HubInterface. SendMessage delegates to a func field
// assigned at runtime; GetIncomingMessagesChannel is backed by a real
// channel that FeedIncomingMessage writes to, so a test can push
// ClientMessage values as if they arrived from a client:
//
//	mock := NewHubMock()
//	mock.SendMessageFunc = func(msg ClientMessage) { sent = append(sent, msg) }
//	go handler.Handle(mock.GetIncomingMessagesChannel())
//	mock.FeedIncomingMessage(ClientMessage{ClientId: "d1"})
type HubMock struct {
	SendMessageFunc func(message ClientMessage)

	IncomingMessagesChannel chan ClientMessage
}

func NewHubMock() *HubMock {
	return &HubMock{
		IncomingMessagesChannel: make(chan ClientMessage, 256),
	}
}

func (m *HubMock) SendMessage(message ClientMessage) {
	if m.SendMessageFunc != nil {
		m.SendMessageFunc(message)
	}
}

func (m *HubMock) GetIncomingMessagesChannel() <-chan ClientMessage {
	return m.IncomingMessagesChannel
}

// FeedIncomingMessage pushes message onto IncomingMessagesChannel, as if it
// had arrived from a client, for a test to observe via the handler under test.
func (m *HubMock) FeedIncomingMessage(message ClientMessage) {
	m.IncomingMessagesChannel <- message
}

// Close closes IncomingMessagesChannel. A handler's Handle ranges over the
// channel, so once every fed message has been consumed, a closed channel
// makes Handle return instead of blocking — letting a test call Handle
// directly (no goroutine, no waitFor) and assert MainState right after.
func (m *HubMock) Close() {
	close(m.IncomingMessagesChannel)
}

var _ HubInterface = (*HubMock)(nil)
