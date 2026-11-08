package state

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

var testNow = time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC)

func newTestRegistry(t *testing.T) *ClientsRegistry {
	t.Helper()

	registry := NewClientsRegistry()
	err := registry.AddSession(ClientSession{ClientId: "client-1", Token: "secret", Tags: []string{"user-42", "room-a", "user-42"}}, testNow)
	if err != nil {
		t.Fatalf("add session: %v", err)
	}

	return registry
}

func TestAddSessionValidation(t *testing.T) {
	registry := newTestRegistry(t)

	if err := registry.AddSession(ClientSession{ClientId: "", Token: "t"}, testNow); !errors.Is(err, ErrEmptyClientId) {
		t.Errorf("empty client id: got %v", err)
	}
	if err := registry.AddSession(ClientSession{ClientId: "c", Token: ""}, testNow); !errors.Is(err, ErrEmptyToken) {
		t.Errorf("empty token: got %v", err)
	}
	if err := registry.AddSession(ClientSession{ClientId: "client-1", Token: "other"}, testNow); !errors.Is(err, ErrClientIdTaken) {
		t.Errorf("duplicate client id: got %v", err)
	}
}

func TestErrorsCarryContext(t *testing.T) {
	registry := newTestRegistry(t)
	_ = registry.AddSession(ClientSession{ClientId: "client-2", Token: "secret2"}, testNow)
	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)

	tests := []struct {
		name     string
		err      error
		wantIs   error
		contains []string
	}{
		{
			name:     "empty token",
			err:      registry.AddSession(ClientSession{ClientId: "client-3"}, testNow),
			wantIs:   ErrEmptyToken,
			contains: []string{"client-3"},
		},
		{
			name:     "client id taken",
			err:      registry.AddSession(ClientSession{ClientId: "client-1", Token: "x"}, testNow),
			wantIs:   ErrClientIdTaken,
			contains: []string{"client-1"},
		},
		{
			name:     "session not found",
			err:      loginError(registry, "nobody", "secret", "conn-9"),
			wantIs:   ErrSessionNotFound,
			contains: []string{"nobody"},
		},
		{
			name:     "invalid token",
			err:      loginError(registry, "client-2", "wrong-token", "conn-9"),
			wantIs:   ErrInvalidToken,
			contains: []string{"client-2"},
		},
		{
			name:     "connection already logged in",
			err:      loginError(registry, "client-2", "secret2", "conn-1"),
			wantIs:   ErrConnectionAlreadyLoggedIn,
			contains: []string{"conn-1", "client-2", "client-1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !errors.Is(test.err, test.wantIs) {
				t.Fatalf("want %v, got %v", test.wantIs, test.err)
			}
			for _, part := range test.contains {
				if !strings.Contains(test.err.Error(), part) {
					t.Errorf("error %q does not mention %q", test.err, part)
				}
			}
		})
	}

	if err := loginError(registry, "client-2", "wrong-token", "conn-9"); strings.Contains(err.Error(), "wrong-token") {
		t.Errorf("error leaks the token: %q", err)
	}
}

func loginError(registry *ClientsRegistry, clientId, token, connectionId string) error {
	_, err := registry.LoginSession(clientId, token, connectionId, testNow)

	return err
}

func TestAddedSessionIsFoundWithoutConnection(t *testing.T) {
	registry := newTestRegistry(t)

	session, ok := registry.FindSessionByClientId("client-1")
	if !ok {
		t.Fatal("session not found")
	}
	if session.ConnectionId != "" || !session.LoggedInAt.IsZero() {
		t.Errorf("session must not be logged in: %+v", session)
	}
	if session.CreatedAt.IsZero() {
		t.Error("CreatedAt must be set")
	}
	if len(session.Tags) != 2 {
		t.Errorf("tags must be deduplicated, got %v", session.Tags)
	}
	if _, ok := registry.FindSessionByConnectionId("conn-1"); ok {
		t.Error("no connection must be found before login")
	}
}

func TestLogin(t *testing.T) {
	registry := newTestRegistry(t)

	if _, err := registry.LoginSession("client-1", "wrong", "conn-1", testNow); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("wrong token: got %v", err)
	}
	if _, err := registry.LoginSession("nobody", "secret", "conn-1", testNow); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("unknown client: got %v", err)
	}
	if _, err := registry.LoginSession("client-1", "secret", "", testNow); !errors.Is(err, ErrEmptyConnectionId) {
		t.Errorf("empty connection id: got %v", err)
	}

	session, err := registry.LoginSession("client-1", "secret", "conn-1", testNow)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if session.ConnectionId != "conn-1" || session.LoggedInAt.IsZero() {
		t.Errorf("session not logged in: %+v", session)
	}

	byConnection, ok := registry.FindSessionByConnectionId("conn-1")
	if !ok || byConnection.ClientId != "client-1" {
		t.Errorf("find by connection: %+v %v", byConnection, ok)
	}

	if _, err := registry.LoginSession("client-1", "secret", "conn-1", testNow); err != nil {
		t.Errorf("login again on the same connection: %v", err)
	}
}

func TestLoginOnConnectionBoundToAnotherClient(t *testing.T) {
	registry := newTestRegistry(t)
	_ = registry.AddSession(ClientSession{ClientId: "client-2", Token: "secret2"}, testNow)
	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)

	if _, err := registry.LoginSession("client-2", "secret2", "conn-1", testNow); !errors.Is(err, ErrConnectionAlreadyLoggedIn) {
		t.Errorf("got %v", err)
	}
}

func TestSecondLoginIsRefusedWhileTheFirstConnectionIsAlive(t *testing.T) {
	registry := newTestRegistry(t)
	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)

	_, err := registry.LoginSession("client-1", "secret", "conn-2", testNow)

	if !errors.Is(err, ErrSessionAlreadyConnected) {
		t.Fatalf("the second login must be refused: got %v", err)
	}
	if _, ok := registry.FindSessionByConnectionId("conn-1"); !ok {
		t.Error("the first connection must stay bound")
	}
	if _, ok := registry.FindSessionByConnectionId("conn-2"); ok {
		t.Error("the refused connection must not be bound")
	}
}

func TestWrongTokenDoesNotTellThatTheSessionIsConnected(t *testing.T) {
	registry := newTestRegistry(t)
	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)

	_, err := registry.LoginSession("client-1", "wrong", "conn-2", testNow)

	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("a wrong token must be reported as a wrong token: got %v", err)
	}
}

func TestFindSessionsByTagReturnsOnlyLoggedIn(t *testing.T) {
	registry := newTestRegistry(t)
	_ = registry.AddSession(ClientSession{ClientId: "client-2", Token: "secret2", Tags: []string{"user-42"}}, testNow)

	if sessions := registry.FindSessionsByTag("user-42"); len(sessions) != 0 {
		t.Fatalf("no one is logged in yet, got %d", len(sessions))
	}

	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)
	sessions := registry.FindSessionsByTag("user-42")
	if len(sessions) != 1 || sessions[0].ClientId != "client-1" {
		t.Fatalf("got %+v", sessions)
	}
	if sessions := registry.FindSessionsByTag("unknown"); len(sessions) != 0 {
		t.Errorf("unknown tag: got %d", len(sessions))
	}
}

func TestRemoveCleansAllIndexes(t *testing.T) {
	registry := newTestRegistry(t)
	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)

	removed, ok := registry.RemoveSessionByConnectionId("conn-1")
	if !ok || removed.ClientId != "client-1" {
		t.Fatalf("remove by connection: %+v %v", removed, ok)
	}
	if _, ok := registry.FindSessionByClientId("client-1"); ok {
		t.Error("still found by client id")
	}
	if len(registry.tag2SessionsMap) != 0 {
		t.Errorf("tag index not cleaned: %v", registry.tag2SessionsMap)
	}

	// a client id can be provisioned again afterwards
	if err := registry.AddSession(ClientSession{ClientId: "client-1", Token: "new"}, testNow); err != nil {
		t.Fatalf("re-add: %v", err)
	}
	if _, ok := registry.RemoveSessionByClientId("client-1"); !ok {
		t.Error("remove by client id failed")
	}
	if _, ok := registry.RemoveSessionByClientId("client-1"); ok {
		t.Error("second remove must find nothing")
	}
}

func TestDisconnectKeepsSessionForReconnect(t *testing.T) {
	registry := newTestRegistry(t)

	if _, ok := registry.DisconnectSession("client-1"); ok {
		t.Error("a session that is not logged in has nothing to disconnect")
	}
	if _, ok := registry.DisconnectSession("nobody"); ok {
		t.Error("unknown client has nothing to disconnect")
	}

	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)

	connectionId, ok := registry.DisconnectSession("client-1")
	if !ok || connectionId != "conn-1" {
		t.Fatalf("disconnect: %q %v", connectionId, ok)
	}
	if _, ok := registry.FindSessionByConnectionId("conn-1"); ok {
		t.Error("connection still found after disconnect")
	}
	if sessions := registry.FindSessionsByTag("user-42"); len(sessions) != 0 {
		t.Errorf("disconnected session still receives tag messages: %+v", sessions)
	}

	session, ok := registry.FindSessionByClientId("client-1")
	if !ok || session.ConnectionId != "" || !session.LoggedInAt.IsZero() || len(session.Tags) != 2 {
		t.Fatalf("session must be kept without a connection: %+v %v", session, ok)
	}

	if _, ok := registry.DisconnectSession("client-1"); ok {
		t.Error("second disconnect must find nothing")
	}
	if _, err := registry.LoginSession("client-1", "secret", "conn-2", testNow); err != nil {
		t.Errorf("login after reconnect: %v", err)
	}
}

func TestDisconnectByConnectionId(t *testing.T) {
	registry := newTestRegistry(t)

	if _, ok := registry.DisconnectSessionByConnectionId("conn-1"); ok {
		t.Error("unknown connection has nothing to disconnect")
	}

	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)

	session, ok := registry.DisconnectSessionByConnectionId("conn-1")
	if !ok || session.ClientId != "client-1" || session.ConnectionId != "" {
		t.Fatalf("disconnect: %+v %v", session, ok)
	}
	if _, ok := registry.FindSessionByConnectionId("conn-1"); ok {
		t.Error("connection still found after disconnect")
	}
	if kept, ok := registry.FindSessionByClientId("client-1"); !ok || kept.ConnectionId != "" {
		t.Errorf("session must be kept without a connection: %+v %v", kept, ok)
	}
	if _, ok := registry.DisconnectSessionByConnectionId("conn-1"); ok {
		t.Error("second disconnect must find nothing")
	}
}

func TestLateDisconnectOfAnOldConnectionKeepsNewLogin(t *testing.T) {
	registry := newTestRegistry(t)
	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)
	_, _ = registry.DisconnectSessionByConnectionId("conn-1")
	_, _ = registry.LoginSession("client-1", "secret", "conn-2", testNow)

	if _, ok := registry.DisconnectSessionByConnectionId("conn-1"); ok {
		t.Error("late close of the old connection detached something")
	}
	session, ok := registry.FindSessionByConnectionId("conn-2")
	if !ok || session.ClientId != "client-1" {
		t.Errorf("new login lost: %+v %v", session, ok)
	}
}

func TestTimesComeFromTheGivenNow(t *testing.T) {
	registry := newTestRegistry(t)

	session, _ := registry.FindSessionByClientId("client-1")
	if !session.CreatedAt.Equal(testNow) {
		t.Errorf("CreatedAt: want %v, got %v", testNow, session.CreatedAt)
	}

	loginTime := testNow.Add(time.Minute)
	session, err := registry.LoginSession("client-1", "secret", "conn-1", loginTime)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if !session.LoggedInAt.Equal(loginTime) || !session.LastActivityAt.Equal(loginTime) {
		t.Errorf("login times: %+v", session)
	}

	// a CreatedAt given by the caller is kept
	provisionedAt := testNow.Add(-time.Hour)
	_ = registry.AddSession(ClientSession{ClientId: "client-2", Token: "t", CreatedAt: provisionedAt}, testNow)
	other, _ := registry.FindSessionByClientId("client-2")
	if !other.CreatedAt.Equal(provisionedAt) {
		t.Errorf("given CreatedAt was overwritten: %v", other.CreatedAt)
	}
}

func TestUpdateSession(t *testing.T) {
	registry := newTestRegistry(t)

	touch := func(session *ClientSession) { session.LastActivityAt = testNow.Add(time.Hour) }
	if registry.UpdateSessionByConnectionId("conn-1", touch) {
		t.Error("a connection that is not logged in has no session to update")
	}
	if registry.UpdateSessionByClientId("nobody", touch) {
		t.Error("unknown client has no session to update")
	}

	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)

	if !registry.UpdateSessionByConnectionId("conn-1", touch) {
		t.Fatal("update by connection failed")
	}
	session, _ := registry.FindSessionByClientId("client-1")
	if !session.LastActivityAt.Equal(testNow.Add(time.Hour)) {
		t.Errorf("update was not stored: %+v", session)
	}

	// the same object is reached by both keys, and the token can be rotated
	if !registry.UpdateSessionByClientId("client-1", func(session *ClientSession) { session.Token = "rotated" }) {
		t.Fatal("update by client id failed")
	}
	byConnection, _ := registry.FindSessionByConnectionId("conn-1")
	if byConnection.Token != "rotated" {
		t.Errorf("token not updated: %+v", byConnection)
	}
	registry.DisconnectSession("client-1")
	if _, err := registry.LoginSession("client-1", "secret", "conn-2", testNow); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("old token must stop working: %v", err)
	}
	if _, err := registry.LoginSession("client-1", "rotated", "conn-2", testNow); err != nil {
		t.Errorf("new token must work: %v", err)
	}
}

func TestTouchSession(t *testing.T) {
	registry := newTestRegistry(t)

	if registry.TouchSession("conn-1", testNow) {
		t.Error("a connection that is not logged in has nothing to touch")
	}

	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)
	later := testNow.Add(5 * time.Minute)
	if !registry.TouchSession("conn-1", later) {
		t.Fatal("touch failed")
	}

	session, _ := registry.FindSessionByConnectionId("conn-1")
	if !session.LastActivityAt.Equal(later) || !session.LoggedInAt.Equal(testNow) {
		t.Errorf("only LastActivityAt must change: %+v", session)
	}
}

func TestSweepExpired(t *testing.T) {
	registry := NewClientsRegistry()
	minutesAgo := func(minutes int) time.Time { return testNow.Add(-time.Duration(minutes) * time.Minute) }
	add := func(clientId string, createdAt time.Time, tags ...string) {
		t.Helper()
		if err := registry.AddSession(ClientSession{ClientId: clientId, Token: "t", Tags: tags, CreatedAt: createdAt}, testNow); err != nil {
			t.Fatalf("add %s: %v", clientId, err)
		}
	}
	login := func(clientId string, loggedInAt time.Time) {
		t.Helper()
		if _, err := registry.LoginSession(clientId, "t", "conn-"+clientId, loggedInAt); err != nil {
			t.Fatalf("login %s: %v", clientId, err)
		}
	}

	add("unconnected-old", minutesAgo(10), "room")
	add("unconnected-new", minutesAgo(1))
	add("idle", minutesAgo(60), "room")
	login("idle", minutesAgo(20))
	add("active", minutesAgo(60))
	login("active", minutesAgo(20))
	registry.TouchSession("conn-active", minutesAgo(1))

	bounds := SweepBounds{
		UnconnectedRegistrationOldestAt: minutesAgo(5),
		LastActionOldestAt:              minutesAgo(10),
	}

	if swept := registry.SweepExpired(SweepBounds{}); len(swept) != 0 {
		t.Fatalf("zero bounds must switch every rule off: %+v", swept)
	}

	swept := registry.SweepExpired(bounds)

	connectionIds := map[string]string{}
	var clientIds []string
	for _, session := range swept {
		clientIds = append(clientIds, session.ClientId)
		connectionIds[session.ClientId] = session.ConnectionId
	}
	slices.Sort(clientIds)

	if want := []string{"idle", "unconnected-old"}; !slices.Equal(clientIds, want) {
		t.Fatalf("swept %v, want %v", clientIds, want)
	}
	// the caller closes the connections of the swept sessions
	if connectionIds["idle"] != "conn-idle" || connectionIds["unconnected-old"] != "" {
		t.Errorf("connection ids in the report: %v", connectionIds)
	}

	for _, clientId := range clientIds {
		if _, ok := registry.FindSessionByClientId(clientId); ok {
			t.Errorf("%s is still found by client id", clientId)
		}
		if _, ok := registry.FindSessionByConnectionId("conn-" + clientId); ok {
			t.Errorf("%s is still found by connection id", clientId)
		}
	}
	if len(registry.tag2SessionsMap) != 0 {
		t.Errorf("tag index not cleaned: %v", registry.tag2SessionsMap)
	}

	// kept: an old CreatedAt does not matter for a logged in session
	for _, clientId := range []string{"unconnected-new", "active"} {
		if _, ok := registry.FindSessionByClientId(clientId); !ok {
			t.Errorf("%s must stay", clientId)
		}
	}

	if again := registry.SweepExpired(bounds); len(again) != 0 {
		t.Errorf("second sweep found %+v", again)
	}
}

// meant to be run with -race: finds share the read lock, updates need the write lock
func TestConcurrentFindAndUpdate(t *testing.T) {
	registry := newTestRegistry(t)
	_, _ = registry.LoginSession("client-1", "secret", "conn-1", testNow)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				registry.FindSessionByConnectionId("conn-1")
				registry.FindSessionsByTag("user-42")
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				registry.UpdateSessionByConnectionId("conn-1", func(session *ClientSession) {
					session.LastActivityAt = testNow.Add(time.Duration(j) * time.Second)
				})
			}
		}()
	}
	wg.Wait()
}

func TestReturnedSessionsAreCopies(t *testing.T) {
	registry := newTestRegistry(t)

	session, _ := registry.FindSessionByClientId("client-1")
	session.Tags[0] = "changed"
	session.ClientId = "changed"

	again, _ := registry.FindSessionByClientId("client-1")
	if again.ClientId != "client-1" || again.Tags[0] != "user-42" {
		t.Errorf("registry state was modified through a returned copy: %+v", again)
	}
}
