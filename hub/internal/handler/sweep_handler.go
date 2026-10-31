package handler

import (
	"log/slog"
	"time"

	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
)

// SweepConfig sets how often the sweeps run and how long things may wait. A
// zero or negative duration switches that part off: an interval without a
// ticker, a timeout without a rule.
type SweepConfig struct {
	// PendingInterval is how often PendingConnections is swept.
	PendingInterval time.Duration
	// PendingConnectionTimeout is how long a connection may wait for its login.
	PendingConnectionTimeout time.Duration

	// SessionInterval is how often ClientsRegistry is swept.
	SessionInterval time.Duration
	// UnconnectedRegistrationTimeout is how long a registered session may wait for its first login.
	UnconnectedRegistrationTimeout time.Duration
	// SessionIdleTimeout is how long a logged in session may stay without activity.
	SessionIdleTimeout time.Duration
}

// SweepHandler closes the connections that wait too long for their login, and
// removes the sessions that never logged in or have been idle too long, closing
// their connections too. The two sweeps run on their own tickers in Run, and
// can be called directly.
type SweepHandler struct {
	pendingConnections *state.PendingConnections
	clientsRegistry    *state.ClientsRegistry
	hub                ws.HubInterface
	clock              util.Clock
	config             SweepConfig
	logger             *slog.Logger
}

func NewSweepHandler(
	pendingConnections *state.PendingConnections,
	clientsRegistry *state.ClientsRegistry,
	hub ws.HubInterface,
	clock util.Clock,
	config SweepConfig,
) *SweepHandler {
	return &SweepHandler{
		pendingConnections: pendingConnections,
		clientsRegistry:    clientsRegistry,
		hub:                hub,
		clock:              clock,
		config:             config,
		logger:             slog.Default().With("handler", "SweepHandler"),
	}
}

// Run sweeps on the tickers until stop is closed. It is meant to run in a goroutine.
func (h *SweepHandler) Run(stop <-chan struct{}) {
	pendingTick, stopPendingTicker := tickerChannel(h.config.PendingInterval)
	defer stopPendingTicker()
	sessionTick, stopSessionTicker := tickerChannel(h.config.SessionInterval)
	defer stopSessionTicker()

	for {
		select {
		case <-pendingTick:
			h.SweepPendingConnections()
		case <-sessionTick:
			h.SweepSessions()
		case <-stop:
			return
		}
	}
}

// SweepPendingConnections closes the connections that did not log in in time.
// A connection that is being logged in right now is not found by the sweep,
// because the login takes it out of PendingConnections first.
func (h *SweepHandler) SweepPendingConnections() {
	oldestAt := boundFor(h.clock.Now(), h.config.PendingConnectionTimeout)

	for _, connection := range h.pendingConnections.SweepExpired(oldestAt) {
		h.logger.Info("closing a connection that did not log in in time",
			"connection_id", connection.ConnectionId, "connected_at", connection.ConnectedAt)
		h.hub.CloseConnection(connection.ConnectionId)
	}
}

// SweepSessions removes the sessions that never logged in or have been idle too
// long. Only the idle ones have a connection, which is closed. The client is
// not told why: the hub closes the socket right away, so a message queued
// before it could be lost.
func (h *SweepHandler) SweepSessions() {
	now := h.clock.Now()
	bounds := state.SweepBounds{
		UnconnectedRegistrationOldestAt: boundFor(now, h.config.UnconnectedRegistrationTimeout),
		LastActionOldestAt:              boundFor(now, h.config.SessionIdleTimeout),
	}

	for _, session := range h.clientsRegistry.SweepExpired(bounds) {
		if session.ConnectionId == "" {
			h.logger.Info("removed a session that never logged in", "client_id", session.ClientId, "created_at", session.CreatedAt)
			continue
		}

		h.logger.Info("closing the connection of an idle session",
			"client_id", session.ClientId, "connection_id", session.ConnectionId, "last_activity_at", session.LastActivityAt)
		h.hub.CloseConnection(session.ConnectionId)
	}
}

// boundFor returns the oldest time still allowed for the timeout, or a zero
// time, which matches nothing, when the timeout is off.
func boundFor(now time.Time, timeout time.Duration) time.Time {
	if timeout <= 0 {
		return time.Time{}
	}

	return now.Add(-timeout)
}

// tickerChannel returns the channel of a new ticker and the function that stops
// it. For an interval that is not positive the channel is nil, and receiving
// from a nil channel blocks forever, so that sweep never runs.
func tickerChannel(interval time.Duration) (<-chan time.Time, func()) {
	if interval <= 0 {
		return nil, func() {}
	}

	ticker := time.NewTicker(interval)

	return ticker.C, ticker.Stop
}
