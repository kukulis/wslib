package state

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrConnectionAlreadyPending  = errors.New("connection is already pending")
	ErrTooManyPendingConnections = errors.New("too many pending connections from one address")
)

// PendingConnection is a connection that was accepted but has not logged in yet.
type PendingConnection struct {
	ConnectionId string
	RemoteIP     string
	ConnectedAt  time.Time
}

// PendingConnections tracks the connections that are open but have not logged
// in yet, so the ones that wait too long can be closed. It does not know the
// ClientsRegistry: the login step calls Remove first, and a true result means
// the sweep can no longer take that connection.
type PendingConnections struct {
	connections map[string]PendingConnection
	countByIP   map[string]int
	maxPerIP    int
	lock        sync.RWMutex
}

// NewPendingConnections creates the registry. maxPerIP limits the pending
// connections of one remote address; zero means no limit.
func NewPendingConnections(maxPerIP int) *PendingConnections {
	return &PendingConnections{
		connections: make(map[string]PendingConnection),
		countByIP:   make(map[string]int),
		maxPerIP:    maxPerIP,
	}
}

// Add registers a new pending connection. An empty remoteIP is not counted
// against the per-address limit, because unknown addresses must not share one
// bucket. To put back a connection after a failed login, pass the ConnectedAt
// that Remove returned, so repeated attempts cannot extend the deadline.
func (p *PendingConnections) Add(connectionId string, remoteIP string, connectedAt time.Time) error {
	if connectionId == "" {
		return ErrEmptyConnectionId
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.connections[connectionId]; ok {
		return fmt.Errorf("%w: connectionId=%q", ErrConnectionAlreadyPending, connectionId)
	}
	if remoteIP != "" && p.maxPerIP > 0 && p.countByIP[remoteIP] >= p.maxPerIP {
		return fmt.Errorf("%w: remoteIP=%q, limit=%d", ErrTooManyPendingConnections, remoteIP, p.maxPerIP)
	}

	p.connections[connectionId] = PendingConnection{
		ConnectionId: connectionId,
		RemoteIP:     remoteIP,
		ConnectedAt:  connectedAt,
	}
	if remoteIP != "" {
		p.countByIP[remoteIP]++
	}

	return nil
}

// Remove takes the connection out and returns it. Only one caller can win: it
// is the claim that decides between a login and a sweep for the same connection.
func (p *PendingConnections) Remove(connectionId string) (PendingConnection, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	connection, ok := p.connections[connectionId]
	if !ok {
		return PendingConnection{}, false
	}
	p.removeLocked(connection)

	return connection, true
}

// SweepExpired removes the connections that connected before connectedOldestAt
// and returns them, so the caller can close them. A zero bound matches nothing.
func (p *PendingConnections) SweepExpired(connectedOldestAt time.Time) []PendingConnection {
	p.lock.Lock()
	defer p.lock.Unlock()

	var swept []PendingConnection

	// deleting the current entry while ranging over a map is safe in Go
	for _, connection := range p.connections {
		if connection.ConnectedAt.Before(connectedOldestAt) {
			swept = append(swept, connection)
			p.removeLocked(connection)
		}
	}

	return swept
}

func (p *PendingConnections) Count() int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return len(p.connections)
}

// removeLocked must be called with the lock held.
func (p *PendingConnections) removeLocked(connection PendingConnection) {
	delete(p.connections, connection.ConnectionId)

	if connection.RemoteIP == "" {
		return
	}
	if p.countByIP[connection.RemoteIP]--; p.countByIP[connection.RemoteIP] <= 0 {
		delete(p.countByIP, connection.RemoteIP)
	}
}
