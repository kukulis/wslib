package handler

import (
	"slices"
	"testing"
	"time"

	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
)

type sweepFixture struct {
	pending  *state.PendingConnections
	registry *state.ClientsRegistry
	hub      *ws.HubMock
	clock    *util.FixedClock
	closed   []string
	handler  *SweepHandler
}

func newSweepFixture(config SweepConfig) *sweepFixture {
	fixture := &sweepFixture{
		pending:  state.NewPendingConnections(0),
		registry: state.NewClientsRegistry(),
		hub:      ws.NewHubMock(),
		clock:    util.NewFixedClock(connectTestNow),
	}
	fixture.hub.CloseConnectionFunc = func(connectionId string) { fixture.closed = append(fixture.closed, connectionId) }
	fixture.handler = NewSweepHandler(fixture.pending, fixture.registry, fixture.hub, fixture.clock, config)

	return fixture
}

func minutesAgo(minutes int) time.Time {
	return connectTestNow.Add(-time.Duration(minutes) * time.Minute)
}

func (f *sweepFixture) addSession(t *testing.T, clientId string, createdAt time.Time) {
	t.Helper()

	if err := f.registry.AddSession(state.ClientSession{ClientId: clientId, Token: "t", CreatedAt: createdAt}, connectTestNow); err != nil {
		t.Fatalf("add %s: %v", clientId, err)
	}
}

func (f *sweepFixture) login(t *testing.T, clientId string, loggedInAt time.Time) {
	t.Helper()

	if _, err := f.registry.LoginSession(clientId, "t", "conn-"+clientId, loggedInAt); err != nil {
		t.Fatalf("login %s: %v", clientId, err)
	}
}

func TestSweepPendingConnectionsClosesTheOnesWaitingTooLong(t *testing.T) {
	fixture := newSweepFixture(SweepConfig{PendingConnectionTimeout: 5 * time.Minute})
	_ = fixture.pending.Add("old", "", minutesAgo(10))
	_ = fixture.pending.Add("new", "", minutesAgo(1))

	fixture.handler.SweepPendingConnections()

	if !slices.Equal(fixture.closed, []string{"old"}) {
		t.Errorf("closed: %v", fixture.closed)
	}
	if fixture.pending.Count() != 1 {
		t.Errorf("pending count: %d", fixture.pending.Count())
	}

	fixture.handler.SweepPendingConnections()
	if len(fixture.closed) != 1 {
		t.Errorf("the second sweep closed again: %v", fixture.closed)
	}
}

func TestSweepPendingConnectionsOffWithoutTimeout(t *testing.T) {
	fixture := newSweepFixture(SweepConfig{})
	_ = fixture.pending.Add("old", "", minutesAgo(1000))

	fixture.handler.SweepPendingConnections()

	if len(fixture.closed) != 0 || fixture.pending.Count() != 1 {
		t.Errorf("closed %v, pending %d", fixture.closed, fixture.pending.Count())
	}
}

func TestSweepSessions(t *testing.T) {
	fixture := newSweepFixture(SweepConfig{
		UnconnectedRegistrationTimeout: 5 * time.Minute,
		SessionIdleTimeout:             10 * time.Minute,
	})
	fixture.addSession(t, "never-logged-in", minutesAgo(10))
	fixture.addSession(t, "fresh", minutesAgo(1))
	fixture.addSession(t, "idle", minutesAgo(60))
	fixture.login(t, "idle", minutesAgo(20))
	fixture.addSession(t, "active", minutesAgo(60))
	fixture.login(t, "active", minutesAgo(20))
	fixture.registry.TouchSession("conn-active", minutesAgo(1))

	fixture.handler.SweepSessions()

	// only the idle session has a connection to close
	if !slices.Equal(fixture.closed, []string{"conn-idle"}) {
		t.Errorf("closed: %v", fixture.closed)
	}
	for _, clientId := range []string{"never-logged-in", "idle"} {
		if _, ok := fixture.registry.FindSessionByClientId(clientId); ok {
			t.Errorf("%s must be removed", clientId)
		}
	}
	for _, clientId := range []string{"fresh", "active"} {
		if _, ok := fixture.registry.FindSessionByClientId(clientId); !ok {
			t.Errorf("%s must stay", clientId)
		}
	}
}

func TestSweepSessionsRulesAreIndependent(t *testing.T) {
	fixture := newSweepFixture(SweepConfig{SessionIdleTimeout: 10 * time.Minute})
	fixture.addSession(t, "never-logged-in", minutesAgo(1000))
	fixture.addSession(t, "idle", minutesAgo(60))
	fixture.login(t, "idle", minutesAgo(20))

	fixture.handler.SweepSessions()

	if _, ok := fixture.registry.FindSessionByClientId("never-logged-in"); !ok {
		t.Error("without UnconnectedRegistrationTimeout that rule must be off")
	}
	if !slices.Equal(fixture.closed, []string{"conn-idle"}) {
		t.Errorf("closed: %v", fixture.closed)
	}
}

func TestSweepSessionsOffWithoutTimeouts(t *testing.T) {
	fixture := newSweepFixture(SweepConfig{})
	fixture.addSession(t, "never-logged-in", minutesAgo(1000))
	fixture.addSession(t, "idle", minutesAgo(1000))
	fixture.login(t, "idle", minutesAgo(1000))

	fixture.handler.SweepSessions()

	if len(fixture.closed) != 0 {
		t.Errorf("closed: %v", fixture.closed)
	}
}

func TestRunSweepsOnTheTickersAndStops(t *testing.T) {
	fixture := newSweepFixture(SweepConfig{
		PendingInterval:                5 * time.Millisecond,
		PendingConnectionTimeout:       5 * time.Minute,
		SessionInterval:                5 * time.Millisecond,
		UnconnectedRegistrationTimeout: 5 * time.Minute,
		SessionIdleTimeout:             10 * time.Minute,
	})
	closed := make(chan string, 10)
	fixture.hub.CloseConnectionFunc = func(connectionId string) { closed <- connectionId }

	_ = fixture.pending.Add("old-pending", "", minutesAgo(10))
	fixture.addSession(t, "idle", minutesAgo(60))
	fixture.login(t, "idle", minutesAgo(20))

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		fixture.handler.Run(stop)
		close(done)
	}()

	var got []string
	for len(got) < 2 {
		select {
		case connectionId := <-closed:
			got = append(got, connectionId)
		case <-time.After(2 * time.Second):
			t.Fatalf("the tickers did not sweep, closed so far: %v", got)
		}
	}
	slices.Sort(got)
	if !slices.Equal(got, []string{"conn-idle", "old-pending"}) {
		t.Errorf("closed: %v", got)
	}

	close(stop)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop")
	}
}

func TestRunWithoutIntervalsOnlyWaitsForStop(t *testing.T) {
	fixture := newSweepFixture(SweepConfig{})

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		fixture.handler.Run(stop)
		close(done)
	}()

	close(stop)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop")
	}
}
