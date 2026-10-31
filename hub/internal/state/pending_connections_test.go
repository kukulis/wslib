package state

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPendingAddValidation(t *testing.T) {
	pending := NewPendingConnections(0)

	if err := pending.Add("", "10.0.0.1", testNow); !errors.Is(err, ErrEmptyConnectionId) {
		t.Errorf("empty connection id: got %v", err)
	}

	if err := pending.Add("conn-1", "10.0.0.1", testNow); err != nil {
		t.Fatalf("add: %v", err)
	}
	err := pending.Add("conn-1", "10.0.0.2", testNow)
	if !errors.Is(err, ErrConnectionAlreadyPending) || !strings.Contains(err.Error(), "conn-1") {
		t.Errorf("duplicate: got %v", err)
	}
	if pending.Count() != 1 {
		t.Errorf("count: %d", pending.Count())
	}
}

func TestPendingRemoveIsAClaim(t *testing.T) {
	pending := NewPendingConnections(0)
	_ = pending.Add("conn-1", "10.0.0.1", testNow)

	connection, ok := pending.Remove("conn-1")
	if !ok || connection.ConnectionId != "conn-1" || connection.RemoteIP != "10.0.0.1" || !connection.ConnectedAt.Equal(testNow) {
		t.Fatalf("remove: %+v %v", connection, ok)
	}
	if _, ok := pending.Remove("conn-1"); ok {
		t.Error("the second remove must lose the claim")
	}
	if _, ok := pending.Remove("unknown"); ok {
		t.Error("unknown connection must not be found")
	}
}

func TestPendingOnlyOneConcurrentRemoveWins(t *testing.T) {
	pending := NewPendingConnections(0)
	_ = pending.Add("conn-1", "10.0.0.1", testNow)

	var winners atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := pending.Remove("conn-1"); ok {
				winners.Add(1)
			}
		}()
	}
	wg.Wait()

	if winners.Load() != 1 {
		t.Errorf("winners: %d", winners.Load())
	}
}

func TestPendingLimitPerAddress(t *testing.T) {
	pending := NewPendingConnections(2)

	_ = pending.Add("conn-1", "10.0.0.1", testNow)
	_ = pending.Add("conn-2", "10.0.0.1", testNow)

	err := pending.Add("conn-3", "10.0.0.1", testNow)
	if !errors.Is(err, ErrTooManyPendingConnections) || !strings.Contains(err.Error(), "10.0.0.1") || !strings.Contains(err.Error(), "limit=2") {
		t.Fatalf("third from the same address: got %v", err)
	}
	if err := pending.Add("conn-4", "10.0.0.2", testNow); err != nil {
		t.Errorf("another address must not be affected: %v", err)
	}

	pending.Remove("conn-1")
	if err := pending.Add("conn-3", "10.0.0.1", testNow); err != nil {
		t.Errorf("a freed slot must be usable: %v", err)
	}
}

func TestPendingNoLimitAndUnknownAddress(t *testing.T) {
	unlimited := NewPendingConnections(0)
	for _, id := range []string{"a", "b", "c", "d"} {
		if err := unlimited.Add(id, "10.0.0.1", testNow); err != nil {
			t.Fatalf("zero must mean no limit: %v", err)
		}
	}

	limited := NewPendingConnections(1)
	for _, id := range []string{"a", "b", "c"} {
		if err := limited.Add(id, "", testNow); err != nil {
			t.Fatalf("unknown addresses must not share a bucket: %v", err)
		}
	}
	if len(limited.countByIP) != 0 {
		t.Errorf("unknown address was counted: %v", limited.countByIP)
	}
}

func TestPendingSweepExpired(t *testing.T) {
	pending := NewPendingConnections(1)
	minutesAgo := func(minutes int) time.Time { return testNow.Add(-time.Duration(minutes) * time.Minute) }

	_ = pending.Add("old-1", "10.0.0.1", minutesAgo(10))
	_ = pending.Add("old-2", "10.0.0.2", minutesAgo(6))
	_ = pending.Add("at-bound", "10.0.0.3", minutesAgo(5))
	_ = pending.Add("new", "10.0.0.4", minutesAgo(1))

	if swept := pending.SweepExpired(time.Time{}); len(swept) != 0 {
		t.Fatalf("a zero bound must match nothing: %+v", swept)
	}

	swept := pending.SweepExpired(minutesAgo(5))

	var ids []string
	for _, connection := range swept {
		ids = append(ids, connection.ConnectionId)
	}
	slices.Sort(ids)
	if !slices.Equal(ids, []string{"old-1", "old-2"}) {
		t.Fatalf("swept %v", ids)
	}
	if pending.Count() != 2 {
		t.Errorf("count after sweep: %d", pending.Count())
	}
	if _, ok := pending.Remove("old-1"); ok {
		t.Error("a swept connection must be gone")
	}

	// the limit slots of the swept ones are free again
	if err := pending.Add("again", "10.0.0.1", testNow); err != nil {
		t.Errorf("slot not freed by the sweep: %v", err)
	}
	if again := pending.SweepExpired(minutesAgo(5)); len(again) != 0 {
		t.Errorf("second sweep found %+v", again)
	}
}

func TestPendingFailedLoginKeepsTheOriginalDeadline(t *testing.T) {
	pending := NewPendingConnections(0)
	connectedAt := testNow.Add(-10 * time.Minute)
	_ = pending.Add("conn-1", "10.0.0.1", connectedAt)

	connection, _ := pending.Remove("conn-1")
	// the login failed: put it back with the original time
	if err := pending.Add(connection.ConnectionId, connection.RemoteIP, connection.ConnectedAt); err != nil {
		t.Fatalf("put back: %v", err)
	}

	if swept := pending.SweepExpired(testNow.Add(-5 * time.Minute)); len(swept) != 1 {
		t.Errorf("the deadline must not restart, swept %+v", swept)
	}
}

func TestPendingNoLeftoverCounters(t *testing.T) {
	pending := NewPendingConnections(3)
	_ = pending.Add("conn-1", "10.0.0.1", testNow)
	_ = pending.Add("conn-2", "10.0.0.1", testNow)

	pending.Remove("conn-1")
	pending.SweepExpired(testNow.Add(time.Minute))

	if pending.Count() != 0 || len(pending.countByIP) != 0 {
		t.Errorf("leftovers: %d connections, counters %v", pending.Count(), pending.countByIP)
	}
}
