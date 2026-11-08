package handler

import (
	"testing"
	"time"

	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
)

type loginFixture struct {
	pending  *state.PendingConnections
	registry *state.ClientsRegistry
	hub      *ws.HubMock
	sent     []ws.ClientMessage
	closed   []string
	handler  *LoginHandler
}

func newLoginFixture(t *testing.T) *loginFixture {
	t.Helper()

	fixture := &loginFixture{
		pending:  state.NewPendingConnections(0),
		registry: state.NewClientsRegistry(),
		hub:      ws.NewHubMock(),
	}
	fixture.hub.SendMessageFunc = func(message ws.ClientMessage) { fixture.sent = append(fixture.sent, message) }
	fixture.hub.CloseConnectionFunc = func(connectionId string) { fixture.closed = append(fixture.closed, connectionId) }
	tokens := util.NewSequenceGenerator([]string{"token-2", "token-3", "token-4"})
	fixture.handler = NewLoginHandler(fixture.pending, fixture.registry, fixture.hub, util.NewFixedClock(connectTestNow), tokens)

	if err := fixture.registry.AddSession(state.ClientSession{ClientId: "client-1", Token: "secret", Tags: []string{"room"}}, connectTestNow); err != nil {
		t.Fatalf("add session: %v", err)
	}

	return fixture
}

func (f *loginFixture) login(connectionId, messageId, clientId, token string) {
	f.handler.HandleLogin(connectionId, messageId, &messages.LoginMessage{ClientId: clientId, Token: token})
}

func (f *loginFixture) lastResult(t *testing.T) (ws.ClientMessage, bool) {
	t.Helper()

	last, result := f.lastReply(t)

	return last, result.Success
}

func (f *loginFixture) lastReply(t *testing.T) (ws.ClientMessage, *messages.LoginResultMessage) {
	t.Helper()

	if len(f.sent) == 0 {
		t.Fatal("no reply was sent")
	}
	last := f.sent[len(f.sent)-1]
	result, ok := last.Message.(*messages.LoginResultMessage)
	if !ok {
		t.Fatalf("reply is %T", last.Message)
	}

	return last, result
}

func TestLoginSuccess(t *testing.T) {
	fixture := newLoginFixture(t)
	_ = fixture.pending.Add("conn-1", "", connectTestNow.Add(-time.Minute))

	fixture.login("conn-1", "m1", "client-1", "secret")

	session, ok := fixture.registry.FindSessionByConnectionId("conn-1")
	if !ok || session.ClientId != "client-1" || !session.LoggedInAt.Equal(connectTestNow) {
		t.Fatalf("session: %+v %v", session, ok)
	}
	if fixture.pending.Count() != 0 {
		t.Errorf("a logged in connection must leave the pending ones: %d", fixture.pending.Count())
	}
	reply, success := fixture.lastResult(t)
	if !success || reply.ConnectionId != "conn-1" || reply.MessageId != "m1" {
		t.Errorf("reply: %+v success=%v", reply, success)
	}
	if len(fixture.closed) != 0 {
		t.Errorf("nothing must be closed: %v", fixture.closed)
	}
}

func TestFailedLoginPutsTheConnectionBackWithItsOriginalTime(t *testing.T) {
	fixture := newLoginFixture(t)
	connectedAt := connectTestNow.Add(-10 * time.Minute)
	_ = fixture.pending.Add("conn-1", "", connectedAt)

	tests := []struct{ name, clientId, token string }{
		{"wrong token", "client-1", "wrong"},
		{"unknown client", "nobody", "secret"},
		{"empty message", "", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture.login("conn-1", "m1", test.clientId, test.token)

			if _, success := fixture.lastResult(t); success {
				t.Error("login must fail")
			}
			if _, ok := fixture.registry.FindSessionByConnectionId("conn-1"); ok {
				t.Error("a failed login must not bind the connection")
			}
			if len(fixture.closed) != 0 {
				t.Errorf("a failed login must keep the connection open: %v", fixture.closed)
			}
		})
	}

	// still pending, and the deadline did not restart
	swept := fixture.pending.SweepExpired(connectTestNow.Add(-5 * time.Minute))
	if len(swept) != 1 || !swept[0].ConnectedAt.Equal(connectedAt) {
		t.Errorf("swept after failed attempts: %+v", swept)
	}
}

func TestLoginCanBeRetriedAfterAFailure(t *testing.T) {
	fixture := newLoginFixture(t)
	_ = fixture.pending.Add("conn-1", "", connectTestNow)

	fixture.login("conn-1", "m1", "client-1", "wrong")
	fixture.login("conn-1", "m2", "client-1", "secret")

	if _, ok := fixture.registry.FindSessionByConnectionId("conn-1"); !ok {
		t.Error("the second attempt must log in")
	}
	if reply, success := fixture.lastResult(t); !success || reply.MessageId != "m2" {
		t.Errorf("reply: %+v success=%v", reply, success)
	}
}

func TestLoginIsRefusedWhenTheConnectionIsNotPending(t *testing.T) {
	fixture := newLoginFixture(t)
	// the sweep took the connection first, or it never registered

	fixture.login("conn-1", "m1", "client-1", "secret")

	if _, success := fixture.lastResult(t); success {
		t.Error("login must be refused")
	}
	if _, ok := fixture.registry.FindSessionByClientId("client-1"); !ok {
		t.Fatal("the session must stay provisioned")
	}
	if session, _ := fixture.registry.FindSessionByClientId("client-1"); session.ConnectionId != "" {
		t.Errorf("a connection that is not pending was logged in: %+v", session)
	}
	if fixture.pending.Count() != 0 {
		t.Errorf("a refused login must not add a pending connection: %d", fixture.pending.Count())
	}
}

func TestLoginOnASessionWithALiveConnectionIsRefused(t *testing.T) {
	fixture := newLoginFixture(t)
	connectedAt := connectTestNow.Add(-time.Minute)
	_ = fixture.pending.Add("conn-1", "", connectTestNow)
	_ = fixture.pending.Add("conn-2", "", connectedAt)

	fixture.login("conn-1", "m1", "client-1", "secret")
	if _, success := fixture.lastResult(t); !success {
		t.Fatal("the first login must succeed")
	}
	// a third party uses the token from the page source, then the rotated one
	fixture.login("conn-2", "m2", "client-1", "secret")
	if _, success := fixture.lastResult(t); success {
		t.Error("the token from the page source must not work after the real login")
	}
	fixture.login("conn-2", "m3", "client-1", "token-2")
	if _, success := fixture.lastResult(t); success {
		t.Error("a login on a session with a live connection must be refused even with the current token")
	}

	if len(fixture.closed) != 0 {
		t.Errorf("the real connection must not be closed: %v", fixture.closed)
	}
	if _, ok := fixture.registry.FindSessionByConnectionId("conn-1"); !ok {
		t.Error("the real connection must stay bound")
	}
	if _, ok := fixture.registry.FindSessionByConnectionId("conn-2"); ok {
		t.Error("the refused connection must not be bound")
	}
	swept := fixture.pending.SweepExpired(connectTestNow.Add(-30 * time.Second))
	if len(swept) != 1 || swept[0].ConnectionId != "conn-2" {
		t.Errorf("the refused connection must be back among the pending ones: %+v", swept)
	}
}

func TestSuccessfulLoginReplacesTheTokenAndSendsTheNewOne(t *testing.T) {
	fixture := newLoginFixture(t)
	_ = fixture.pending.Add("conn-1", "", connectTestNow)

	fixture.login("conn-1", "m1", "client-1", "secret")

	_, result := fixture.lastReply(t)
	if !result.Success {
		t.Fatal("the login must succeed")
	}
	if result.Token != "token-2" {
		t.Errorf("the reply must carry the new token: %q", result.Token)
	}
	session, _ := fixture.registry.FindSessionByClientId("client-1")
	if session.Token != "token-2" {
		t.Errorf("the session must keep the new token: %q", session.Token)
	}
}

func TestFailedLoginSendsNoToken(t *testing.T) {
	fixture := newLoginFixture(t)
	_ = fixture.pending.Add("conn-1", "", connectTestNow)

	fixture.login("conn-1", "m1", "client-1", "wrong")

	_, result := fixture.lastReply(t)
	if result.Success {
		t.Fatal("the login must fail")
	}
	if result.Token != "" {
		t.Errorf("a failed login must not send a token: %q", result.Token)
	}
}

func TestReconnectNeedsTheRotatedToken(t *testing.T) {
	fixture := newLoginFixture(t)
	_ = fixture.pending.Add("conn-1", "", connectTestNow)
	fixture.login("conn-1", "m1", "client-1", "secret")
	_, firstResult := fixture.lastReply(t)
	if !firstResult.Success {
		t.Fatal("the first login must succeed")
	}

	// the socket closes, and the browser connects again
	fixture.registry.DisconnectSessionByConnectionId("conn-1")
	_ = fixture.pending.Add("conn-2", "", connectTestNow)

	fixture.login("conn-2", "m2", "client-1", "secret")
	if _, success := fixture.lastResult(t); success {
		t.Fatal("the token from the page source must not work for the reconnect")
	}
	fixture.login("conn-2", "m3", "client-1", firstResult.Token)
	_, secondResult := fixture.lastReply(t)
	if !secondResult.Success {
		t.Fatal("the reconnect with the rotated token must succeed")
	}
	if secondResult.Token == firstResult.Token {
		t.Errorf("the token must change on every login: %q", secondResult.Token)
	}
}

func TestLoginHandlerIgnoresOtherMessageTypes(t *testing.T) {
	fixture := newLoginFixture(t)
	_ = fixture.pending.Add("conn-1", "", connectTestNow)

	fixture.handler.HandleLogin("conn-1", "m1", messages.NewDefaultMessage("something", "m1"))

	if len(fixture.sent) != 0 || fixture.pending.Count() != 1 {
		t.Errorf("sent %v, pending %d", fixture.sent, fixture.pending.Count())
	}
}

func TestLoginHandlerWiredWithDelegatingMessageHandler(t *testing.T) {
	fixture := newLoginFixture(t)
	_ = fixture.pending.Add("conn-1", "", connectTestNow)

	delegating := NewDelegatingMessageHandler()
	delegating.SetHandler(messages.MessageTypeLogin, fixture.handler.HandleLogin)

	fixture.hub.FeedIncomingMessage(ws.ClientMessage{
		ConnectionId: "conn-1",
		MessageId:    "m1",
		Message:      &messages.LoginMessage{ClientId: "client-1", Token: "secret"},
	})
	fixture.hub.Close()
	delegating.Handle(fixture.hub.GetIncomingMessagesChannel())

	if _, ok := fixture.registry.FindSessionByConnectionId("conn-1"); !ok {
		t.Error("the login message did not reach the handler")
	}
}
