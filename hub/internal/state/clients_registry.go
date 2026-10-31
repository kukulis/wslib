package state

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"darbelis.eu/wshub/pkg/util"
)

// Callers match these with errors.Is; the registry wraps them with the ids involved.
var (
	ErrEmptyClientId             = errors.New("client id is empty")
	ErrEmptyConnectionId         = errors.New("connection id is empty")
	ErrEmptyToken                = errors.New("token is empty")
	ErrClientIdTaken             = errors.New("client id already has a session")
	ErrSessionNotFound           = errors.New("session not found")
	ErrInvalidToken              = errors.New("invalid token")
	ErrConnectionAlreadyLoggedIn = errors.New("connection is already logged in as another client")
)

type ClientsRegistry struct {
	connection2SessionMap map[string]*ClientSession
	client2SessionMap     map[string]*ClientSession
	tag2SessionsMap       map[string][]*ClientSession
	// One lock for all maps: every change touches several of them at once,
	// separate locks would let readers see the maps out of sync.
	clientsRegistryLock sync.Mutex
}

func NewClientsRegistry() *ClientsRegistry {
	return &ClientsRegistry{
		connection2SessionMap: make(map[string]*ClientSession),
		client2SessionMap:     make(map[string]*ClientSession),
		tag2SessionsMap:       make(map[string][]*ClientSession),
	}
}

// AddSession stores a provisioned session. It stays without a connection until
// LoginSession binds one to it. ConnectionId and LoggedInAt of the argument are
// ignored. The registry keeps its own copy.
func (r *ClientsRegistry) AddSession(session ClientSession) error {
	if session.ClientId == "" {
		return ErrEmptyClientId
	}
	if session.Token == "" {
		return fmt.Errorf("%w: clientId=%q", ErrEmptyToken, session.ClientId)
	}

	session.ConnectionId = ""
	session.LoggedInAt = time.Time{}
	session.Tags = slices.DeleteFunc(util.UniqueSlice(session.Tags), func(tag string) bool {
		return tag == ""
	})
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	r.clientsRegistryLock.Lock()
	defer r.clientsRegistryLock.Unlock()

	if _, ok := r.client2SessionMap[session.ClientId]; ok {
		return fmt.Errorf("%w: clientId=%q", ErrClientIdTaken, session.ClientId)
	}

	stored := &session
	r.client2SessionMap[stored.ClientId] = stored
	for _, tag := range stored.Tags {
		r.tag2SessionsMap[tag] = append(r.tag2SessionsMap[tag], stored)
	}

	return nil
}

// FindSessionByClientId returns a copy of the session, logged in or not.
func (r *ClientsRegistry) FindSessionByClientId(clientId string) (ClientSession, bool) {
	r.clientsRegistryLock.Lock()
	defer r.clientsRegistryLock.Unlock()

	session, ok := r.client2SessionMap[clientId]
	if !ok {
		return ClientSession{}, false
	}

	return snapshot(session), true
}

// FindSessionByConnectionId returns a copy of the logged in session bound to the connection.
func (r *ClientsRegistry) FindSessionByConnectionId(connectionId string) (ClientSession, bool) {
	r.clientsRegistryLock.Lock()
	defer r.clientsRegistryLock.Unlock()

	session, ok := r.connection2SessionMap[connectionId]
	if !ok {
		return ClientSession{}, false
	}

	return snapshot(session), true
}

// LoginSession binds connectionId to the provisioned session of clientId after
// checking the token. If the session was logged in on another connection, that
// connection is replaced and its id is returned, so the caller can close it.
// Logging in again on the same connection changes nothing.
func (r *ClientsRegistry) LoginSession(
	clientId string,
	token string,
	connectionId string,
) (session ClientSession, replacedConnectionId string, err error) {
	if connectionId == "" {
		return ClientSession{}, "", ErrEmptyConnectionId
	}

	r.clientsRegistryLock.Lock()
	defer r.clientsRegistryLock.Unlock()

	stored, ok := r.client2SessionMap[clientId]
	if !ok {
		return ClientSession{}, "", fmt.Errorf("%w: clientId=%q", ErrSessionNotFound, clientId)
	}
	// the token itself must never end up in an error message
	if subtle.ConstantTimeCompare([]byte(stored.Token), []byte(token)) != 1 {
		return ClientSession{}, "", fmt.Errorf("%w: clientId=%q", ErrInvalidToken, clientId)
	}

	if other, bound := r.connection2SessionMap[connectionId]; bound && other != stored {
		return ClientSession{}, "", fmt.Errorf(
			"%w: connectionId=%q, requested clientId=%q, bound clientId=%q",
			ErrConnectionAlreadyLoggedIn, connectionId, clientId, other.ClientId,
		)
	}
	if stored.ConnectionId == connectionId {
		return snapshot(stored), "", nil
	}

	if stored.ConnectionId != "" {
		replacedConnectionId = stored.ConnectionId
		delete(r.connection2SessionMap, replacedConnectionId)
	}

	now := time.Now()
	stored.ConnectionId = connectionId
	stored.LoggedInAt = now
	stored.LastActivityAt = now
	r.connection2SessionMap[connectionId] = stored

	return snapshot(stored), replacedConnectionId, nil
}

// RemoveSessionByConnectionId removes the whole session bound to the connection.
// A connection that was already replaced by a newer login is not found here, so
// its late disconnect cannot remove the newer one.
func (r *ClientsRegistry) RemoveSessionByConnectionId(connectionId string) (ClientSession, bool) {
	r.clientsRegistryLock.Lock()
	defer r.clientsRegistryLock.Unlock()

	session, ok := r.connection2SessionMap[connectionId]
	if !ok {
		return ClientSession{}, false
	}
	r.removeLocked(session)

	return snapshot(session), true
}

// RemoveSessionByClientId removes the session, logged in or not.
func (r *ClientsRegistry) RemoveSessionByClientId(clientId string) (ClientSession, bool) {
	r.clientsRegistryLock.Lock()
	defer r.clientsRegistryLock.Unlock()

	session, ok := r.client2SessionMap[clientId]
	if !ok {
		return ClientSession{}, false
	}
	r.removeLocked(session)

	return snapshot(session), true
}

// DisconnectSession detaches the connection from the session but keeps the
// session, so the same client id and token can log in again after a
// reconnect. It returns the id of the detached connection, and false when the
// session is unknown or not logged in.
func (r *ClientsRegistry) DisconnectSession(clientId string) (string, bool) {
	r.clientsRegistryLock.Lock()
	defer r.clientsRegistryLock.Unlock()

	session, ok := r.client2SessionMap[clientId]
	if !ok || session.ConnectionId == "" {
		return "", false
	}

	return r.disconnectLocked(session), true
}

// DisconnectSessionByConnectionId does the same for the session bound to the
// connection and returns a copy of it, already without the connection. Lookup
// and detach happen under one lock, and a connection that was already replaced
// by a newer login is not found, so a late close of the old socket cannot
// detach the new one. Use this one when a connection closes.
func (r *ClientsRegistry) DisconnectSessionByConnectionId(connectionId string) (ClientSession, bool) {
	r.clientsRegistryLock.Lock()
	defer r.clientsRegistryLock.Unlock()

	session, ok := r.connection2SessionMap[connectionId]
	if !ok {
		return ClientSession{}, false
	}
	r.disconnectLocked(session)

	return snapshot(session), true
}

// FindSessionsByTag returns copies of the logged in sessions carrying the tag;
// sessions without a connection cannot receive anything, so they are left out.
func (r *ClientsRegistry) FindSessionsByTag(tag string) []ClientSession {
	r.clientsRegistryLock.Lock()
	defer r.clientsRegistryLock.Unlock()

	tagged := r.tag2SessionsMap[tag]
	sessions := make([]ClientSession, 0, len(tagged))
	for _, session := range tagged {
		if session.ConnectionId != "" {
			sessions = append(sessions, snapshot(session))
		}
	}

	return sessions
}

// disconnectLocked must be called with clientsRegistryLock held, for a logged in session.
// It returns the id of the detached connection.
func (r *ClientsRegistry) disconnectLocked(session *ClientSession) string {
	connectionId := session.ConnectionId
	delete(r.connection2SessionMap, connectionId)
	session.ConnectionId = ""
	session.LoggedInAt = time.Time{}

	return connectionId
}

// removeLocked must be called with clientsRegistryLock held.
func (r *ClientsRegistry) removeLocked(session *ClientSession) {
	delete(r.client2SessionMap, session.ClientId)
	if session.ConnectionId != "" {
		delete(r.connection2SessionMap, session.ConnectionId)
	}

	for _, tag := range session.Tags {
		remaining := slices.DeleteFunc(r.tag2SessionsMap[tag], func(other *ClientSession) bool {
			return other == session
		})
		if len(remaining) == 0 {
			delete(r.tag2SessionsMap, tag)
		} else {
			r.tag2SessionsMap[tag] = remaining
		}
	}
}

func snapshot(session *ClientSession) ClientSession {
	copied := *session
	copied.Tags = slices.Clone(session.Tags)

	return copied
}
