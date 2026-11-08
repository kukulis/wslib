package handler

import (
	"slices"
	"testing"
	"time"

	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/pkg/message_common"
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
)

var connectTestNow = time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC)

type connectFixture struct {
	pending      *state.PendingConnections
	registry     *state.ClientsRegistry
	hub          *ws.HubMock
	closed       []string
	handler      *ConnectHandler
	waitResponse *WaitResponseMessageHandler
}

func newConnectFixture() *connectFixture {
	fixture := &connectFixture{
		pending:      state.NewPendingConnections(0),
		registry:     state.NewClientsRegistry(),
		hub:          ws.NewHubMock(),
		waitResponse: NewWaitResponseMessageHandler(),
	}
	fixture.hub.CloseConnectionFunc = func(connectionId string) {
		fixture.closed = append(fixture.closed, connectionId)
	}
	fixture.handler = NewConnectHandler(fixture.pending, fixture.registry, fixture.hub, fixture.waitResponse, util.NewFixedClock(connectTestNow))

	return fixture
}

func (f *connectFixture) wire() *DelegatingMessageHandler {
	delegating := NewDelegatingMessageHandler()
	delegating.SetHandler(message_common.VirtualMessageTypeConnectionRegistered, f.handler.HandleConnectionRegistered)
	delegating.SetHandler(message_common.VirtualMessageTypeConnectionUnregistered, f.handler.HandleConnectionUnregistered)

	return delegating
}

func TestConnectionRegisteredBecomesPending(t *testing.T) {
	fixture := newConnectFixture()

	fixture.handler.HandleConnectionRegistered("conn-1", "", &message_common.ConnectionRegisteredMessage{ConnectionId: "conn-1"})

	if fixture.pending.Count() != 1 {
		t.Fatalf("pending count: %d", fixture.pending.Count())
	}
	if len(fixture.closed) != 0 {
		t.Errorf("a tracked connection must stay open: %v", fixture.closed)
	}

	// the connect time comes from the injected clock
	swept := fixture.pending.SweepExpired(connectTestNow.Add(time.Second))
	if len(swept) != 1 || !swept[0].ConnectedAt.Equal(connectTestNow) {
		t.Errorf("swept: %+v", swept)
	}
}

func TestConnectionThatCannotBeTrackedIsClosed(t *testing.T) {
	fixture := newConnectFixture()

	registered := &message_common.ConnectionRegisteredMessage{ConnectionId: "conn-1"}
	fixture.handler.HandleConnectionRegistered("conn-1", "", registered)
	fixture.handler.HandleConnectionRegistered("conn-1", "", registered)

	if !slices.Equal(fixture.closed, []string{"conn-1"}) {
		t.Errorf("closed: %v", fixture.closed)
	}
	if fixture.pending.Count() != 1 {
		t.Errorf("pending count: %d", fixture.pending.Count())
	}
}

func TestConnectionRegisteredCountsTheRemoteAddress(t *testing.T) {
	fixture := newConnectFixture()
	fixture.pending = state.NewPendingConnections(1)
	fixture.handler = NewConnectHandler(fixture.pending, fixture.registry, fixture.hub, fixture.waitResponse, util.NewFixedClock(connectTestNow))

	fixture.handler.HandleConnectionRegistered("conn-1", "", &message_common.ConnectionRegisteredMessage{ConnectionId: "conn-1", RemoteIp: "10.0.0.1"})
	fixture.handler.HandleConnectionRegistered("conn-2", "", &message_common.ConnectionRegisteredMessage{ConnectionId: "conn-2", RemoteIp: "10.0.0.1"})
	fixture.handler.HandleConnectionRegistered("conn-3", "", &message_common.ConnectionRegisteredMessage{ConnectionId: "conn-3", RemoteIp: "10.0.0.2"})

	if !slices.Equal(fixture.closed, []string{"conn-2"}) {
		t.Errorf("only the second connection of one address must be closed: %v", fixture.closed)
	}
}

func TestConnectionRegisteredPanicsOnAnotherMessageType(t *testing.T) {
	fixture := newConnectFixture()

	defer func() {
		if recover() == nil {
			t.Error("a message of another type must be reported as a misconfiguration")
		}
	}()

	// a value is not what the hub sends either: it sends a pointer
	fixture.handler.HandleConnectionRegistered("conn-1", "", message_common.ConnectionRegisteredMessage{ConnectionId: "conn-1"})
}

func TestConnectionUnregisteredRemovesPending(t *testing.T) {
	fixture := newConnectFixture()
	_ = fixture.pending.Add("conn-1", "", connectTestNow)
	_ = fixture.pending.Add("conn-2", "", connectTestNow)

	fixture.handler.HandleConnectionUnregistered("conn-1", "", nil)

	if fixture.pending.Count() != 1 {
		t.Errorf("pending count: %d", fixture.pending.Count())
	}
	if _, ok := fixture.pending.Remove("conn-2"); !ok {
		t.Error("the other connection must stay pending")
	}
}

func TestConnectionUnregisteredDetachesLoggedInSession(t *testing.T) {
	fixture := newConnectFixture()
	_ = fixture.registry.AddSession(state.ClientSession{ClientId: "client-1", Token: "secret"}, connectTestNow)
	_, _ = fixture.registry.LoginSession("client-1", "secret", "conn-1", connectTestNow)

	fixture.handler.HandleConnectionUnregistered("conn-1", "", nil)

	if _, ok := fixture.registry.FindSessionByConnectionId("conn-1"); ok {
		t.Error("the connection is still bound to a session")
	}
	session, ok := fixture.registry.FindSessionByClientId("client-1")
	if !ok || session.ConnectionId != "" {
		t.Errorf("the session must be kept without a connection: %+v %v", session, ok)
	}
	if len(fixture.closed) != 0 {
		t.Errorf("an already closed connection must not be closed again: %v", fixture.closed)
	}
}

func TestConnectionUnregisteredOfUnknownConnectionIsIgnored(t *testing.T) {
	fixture := newConnectFixture()

	fixture.handler.HandleConnectionUnregistered("nobody", "", nil)

	if fixture.pending.Count() != 0 || len(fixture.closed) != 0 {
		t.Errorf("pending %d, closed %v", fixture.pending.Count(), fixture.closed)
	}
}

func TestConnectHandlerWiredWithDelegatingMessageHandler(t *testing.T) {
	fixture := newConnectFixture()
	delegating := fixture.wire()

	fixture.hub.FeedIncomingMessage(ws.ClientMessage{
		ConnectionId: "conn-1",
		Message:      &message_common.ConnectionRegisteredMessage{ConnectionId: "conn-1"},
	})
	fixture.hub.Close()
	delegating.Handle(fixture.hub.GetIncomingMessagesChannel())

	if fixture.pending.Count() != 1 {
		t.Fatalf("registered message did not reach the handler, pending: %d", fixture.pending.Count())
	}

	// a second fixture for the closing side, because the mock channel is closed now
	fixture = newConnectFixture()
	delegating = fixture.wire()
	_ = fixture.pending.Add("conn-1", "", connectTestNow)

	fixture.hub.FeedIncomingMessage(ws.ClientMessage{
		ConnectionId: "conn-1",
		Message:      &message_common.ConnectionUnregisteredMessage{ConnectionId: "conn-1"},
	})
	fixture.hub.Close()
	delegating.Handle(fixture.hub.GetIncomingMessagesChannel())

	if fixture.pending.Count() != 0 {
		t.Errorf("unregistered message did not reach the handler, pending: %d", fixture.pending.Count())
	}
}

func TestConnectionUnregisteredReleasesWaitersOfThatConnectionOnly(t *testing.T) {
	fixture := newConnectFixture()

	closedResult := make(chan message_common.MessageInterface, 1)
	go func() { closedResult <- fixture.waitResponse.WaitForResponse("conn-1", "msg-1", 5*time.Second) }()
	otherResult := make(chan message_common.MessageInterface, 1)
	go func() { otherResult <- fixture.waitResponse.WaitForResponse("conn-2", "msg-2", 300*time.Millisecond) }()
	time.Sleep(20 * time.Millisecond) // let both register

	started := time.Now()
	fixture.handler.HandleConnectionUnregistered("conn-1", "", nil)

	select {
	case got := <-closedResult:
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
		if time.Since(started) > time.Second {
			t.Error("the waiter must be released at once, not after its delay")
		}
	case <-time.After(time.Second):
		t.Fatal("the waiter of the closed connection was not released")
	}

	fixture.waitResponse.lock.Lock()
	_, other := fixture.waitResponse.waiters["msg-2"]
	fixture.waitResponse.lock.Unlock()
	if !other {
		t.Error("the waiter of another connection must stay")
	}
	<-otherResult
}
