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
	pending  *state.PendingConnections
	registry *state.ClientsRegistry
	hub      *ws.HubMock
	closed   []string
	handler  *ConnectHandler
}

func newConnectFixture() *connectFixture {
	fixture := &connectFixture{
		pending:  state.NewPendingConnections(0),
		registry: state.NewClientsRegistry(),
		hub:      ws.NewHubMock(),
	}
	fixture.hub.CloseConnectionFunc = func(connectionId string) {
		fixture.closed = append(fixture.closed, connectionId)
	}
	fixture.handler = NewConnectHandler(fixture.pending, fixture.registry, fixture.hub, util.NewFixedClock(connectTestNow))

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

	fixture.handler.HandleConnectionRegistered("conn-1", "", nil)
	fixture.handler.HandleConnectionRegistered("conn-1", "", nil)

	if !slices.Equal(fixture.closed, []string{"conn-1"}) {
		t.Errorf("closed: %v", fixture.closed)
	}
	if fixture.pending.Count() != 1 {
		t.Errorf("pending count: %d", fixture.pending.Count())
	}
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
	_, _, _ = fixture.registry.LoginSession("client-1", "secret", "conn-1", connectTestNow)

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
