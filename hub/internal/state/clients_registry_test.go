package state

import (
	"errors"
	"strings"
	"testing"
)

func newTestRegistry(t *testing.T) *ClientsRegistry {
	t.Helper()

	registry := NewClientsRegistry()
	err := registry.AddSession(ClientSession{ClientId: "client-1", Token: "secret", Tags: []string{"user-42", "room-a", "user-42"}})
	if err != nil {
		t.Fatalf("add session: %v", err)
	}

	return registry
}

func TestAddSessionValidation(t *testing.T) {
	registry := newTestRegistry(t)

	if err := registry.AddSession(ClientSession{ClientId: "", Token: "t"}); !errors.Is(err, ErrEmptyClientId) {
		t.Errorf("empty client id: got %v", err)
	}
	if err := registry.AddSession(ClientSession{ClientId: "c", Token: ""}); !errors.Is(err, ErrEmptyToken) {
		t.Errorf("empty token: got %v", err)
	}
	if err := registry.AddSession(ClientSession{ClientId: "client-1", Token: "other"}); !errors.Is(err, ErrClientIdTaken) {
		t.Errorf("duplicate client id: got %v", err)
	}
}

func TestErrorsCarryContext(t *testing.T) {
	registry := newTestRegistry(t)
	_ = registry.AddSession(ClientSession{ClientId: "client-2", Token: "secret2"})
	_, _, _ = registry.LoginSession("client-1", "secret", "conn-1")

	tests := []struct {
		name     string
		err      error
		wantIs   error
		contains []string
	}{
		{
			name:     "empty token",
			err:      registry.AddSession(ClientSession{ClientId: "client-3"}),
			wantIs:   ErrEmptyToken,
			contains: []string{"client-3"},
		},
		{
			name:     "client id taken",
			err:      registry.AddSession(ClientSession{ClientId: "client-1", Token: "x"}),
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
	_, _, err := registry.LoginSession(clientId, token, connectionId)

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

	if _, _, err := registry.LoginSession("client-1", "wrong", "conn-1"); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("wrong token: got %v", err)
	}
	if _, _, err := registry.LoginSession("nobody", "secret", "conn-1"); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("unknown client: got %v", err)
	}
	if _, _, err := registry.LoginSession("client-1", "secret", ""); !errors.Is(err, ErrEmptyConnectionId) {
		t.Errorf("empty connection id: got %v", err)
	}

	session, replaced, err := registry.LoginSession("client-1", "secret", "conn-1")
	if err != nil || replaced != "" {
		t.Fatalf("login: replaced=%q err=%v", replaced, err)
	}
	if session.ConnectionId != "conn-1" || session.LoggedInAt.IsZero() {
		t.Errorf("session not logged in: %+v", session)
	}

	byConnection, ok := registry.FindSessionByConnectionId("conn-1")
	if !ok || byConnection.ClientId != "client-1" {
		t.Errorf("find by connection: %+v %v", byConnection, ok)
	}

	if _, replaced, err := registry.LoginSession("client-1", "secret", "conn-1"); err != nil || replaced != "" {
		t.Errorf("login again on the same connection: replaced=%q err=%v", replaced, err)
	}
}

func TestLoginOnConnectionBoundToAnotherClient(t *testing.T) {
	registry := newTestRegistry(t)
	_ = registry.AddSession(ClientSession{ClientId: "client-2", Token: "secret2"})
	_, _, _ = registry.LoginSession("client-1", "secret", "conn-1")

	if _, _, err := registry.LoginSession("client-2", "secret2", "conn-1"); !errors.Is(err, ErrConnectionAlreadyLoggedIn) {
		t.Errorf("got %v", err)
	}
}

func TestSecondLoginReplacesOldConnection(t *testing.T) {
	registry := newTestRegistry(t)
	_, _, _ = registry.LoginSession("client-1", "secret", "conn-1")

	_, replaced, err := registry.LoginSession("client-1", "secret", "conn-2")
	if err != nil || replaced != "conn-1" {
		t.Fatalf("replaced=%q err=%v", replaced, err)
	}
	if _, ok := registry.FindSessionByConnectionId("conn-1"); ok {
		t.Error("old connection must not be found")
	}

	// the late disconnect of the old connection must not remove the new login
	if _, removed := registry.RemoveSessionByConnectionId("conn-1"); removed {
		t.Error("old connection removed the newer session")
	}
	if _, ok := registry.FindSessionByConnectionId("conn-2"); !ok {
		t.Error("new connection lost its session")
	}
}

func TestFindSessionsByTagReturnsOnlyLoggedIn(t *testing.T) {
	registry := newTestRegistry(t)
	_ = registry.AddSession(ClientSession{ClientId: "client-2", Token: "secret2", Tags: []string{"user-42"}})

	if sessions := registry.FindSessionsByTag("user-42"); len(sessions) != 0 {
		t.Fatalf("no one is logged in yet, got %d", len(sessions))
	}

	_, _, _ = registry.LoginSession("client-1", "secret", "conn-1")
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
	_, _, _ = registry.LoginSession("client-1", "secret", "conn-1")

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
	if err := registry.AddSession(ClientSession{ClientId: "client-1", Token: "new"}); err != nil {
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

	_, _, _ = registry.LoginSession("client-1", "secret", "conn-1")

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
	if _, replaced, err := registry.LoginSession("client-1", "secret", "conn-2"); err != nil || replaced != "" {
		t.Errorf("login after reconnect: replaced=%q err=%v", replaced, err)
	}
}

func TestDisconnectByConnectionId(t *testing.T) {
	registry := newTestRegistry(t)

	if _, ok := registry.DisconnectSessionByConnectionId("conn-1"); ok {
		t.Error("unknown connection has nothing to disconnect")
	}

	_, _, _ = registry.LoginSession("client-1", "secret", "conn-1")

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

func TestLateDisconnectOfReplacedConnectionKeepsNewLogin(t *testing.T) {
	registry := newTestRegistry(t)
	_, _, _ = registry.LoginSession("client-1", "secret", "conn-1")
	_, _, _ = registry.LoginSession("client-1", "secret", "conn-2")

	if _, ok := registry.DisconnectSessionByConnectionId("conn-1"); ok {
		t.Error("late close of the replaced connection detached something")
	}
	session, ok := registry.FindSessionByConnectionId("conn-2")
	if !ok || session.ClientId != "client-1" {
		t.Errorf("new login lost: %+v %v", session, ok)
	}
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
