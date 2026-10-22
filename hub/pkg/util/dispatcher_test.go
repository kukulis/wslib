package util

import (
	"testing"
)

type testEvent struct {
	name    string
	payload string
}

func (e *testEvent) GetName() string { return e.name }

func TestDispatcherCallsListener(t *testing.T) {
	dispatcher := NewDispatcher()
	called := false

	dispatcher.AddListener("test", func(e Event) {
		called = true
	})
	dispatcher.Dispatch(&testEvent{name: "test"})

	if !called {
		t.Error("expected listener to be called")
	}
}

func TestDispatcherPassesEvent(t *testing.T) {
	dispatcher := NewDispatcher()
	var received Event

	dispatcher.AddListener("test", func(e Event) {
		received = e
	})
	event := &testEvent{name: "test", payload: "hello"}
	dispatcher.Dispatch(event)

	if received != event {
		t.Error("expected received event to be the dispatched event")
	}
}

func TestDispatcherCallsMultipleListeners(t *testing.T) {
	dispatcher := NewDispatcher()
	count := 0

	dispatcher.AddListener("test", func(e Event) { count++ })
	dispatcher.AddListener("test", func(e Event) { count++ })
	dispatcher.AddListener("test", func(e Event) { count++ })
	dispatcher.Dispatch(&testEvent{name: "test"})

	if count != 3 {
		t.Errorf("expected 3 listeners called, got %d", count)
	}
}

func TestDispatcherOnlyCallsMatchingListeners(t *testing.T) {
	dispatcher := NewDispatcher()
	called := false

	dispatcher.AddListener("other", func(e Event) { called = true })
	dispatcher.Dispatch(&testEvent{name: "test"})

	if called {
		t.Error("expected listener not to be called for a different event name")
	}
}
