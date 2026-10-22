package util

import (
	"slices"
	"sync"
)

type Event interface {
	GetName() string
}

type EventHandler func(Event)

type Dispatcher struct {
	mu        sync.RWMutex
	listeners map[string][]EventHandler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		listeners: make(map[string][]EventHandler),
	}
}

func (d *Dispatcher) AddListener(eventName string, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.listeners[eventName] = append(d.listeners[eventName], handler)
}

func (d *Dispatcher) Dispatch(event Event) {
	d.mu.RLock()
	handlers := slices.Clone(d.listeners[event.GetName()])
	d.mu.RUnlock()
	for _, handler := range handlers {
		handler(event)
	}
}
