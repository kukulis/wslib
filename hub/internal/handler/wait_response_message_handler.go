package handler

import (
	"sync"
	"time"

	"darbelis.eu/wshub/pkg/message_common"
	"github.com/bytedance/gopkg/util/logger"
)

type WaitResponseMessageHandler struct {
	waiters map[string]*Waiter
	lock    sync.Mutex
}

func NewWaitResponseMessageHandler() *WaitResponseMessageHandler {
	return &WaitResponseMessageHandler{
		waiters: make(map[string]*Waiter),
		lock:    sync.Mutex{},
	}
}

func (wmh *WaitResponseMessageHandler) AddWaiter(w *Waiter) bool {
	wmh.lock.Lock()
	defer wmh.lock.Unlock()

	existing := wmh.waiters[w.messageId]
	if existing != nil {
		return false
	}

	wmh.waiters[w.messageId] = w
	return true
}

func (wmh *WaitResponseMessageHandler) RemoveWaiter(w *Waiter) {
	wmh.lock.Lock()
	defer wmh.lock.Unlock()

	// only remove our own registration; a newer waiter may have reused the id
	if wmh.waiters[w.messageId] == w {
		delete(wmh.waiters, w.messageId)
	}
}

// CancelConnection releases every waiter of a closed connection: its channel is
// closed, so WaitForResponse returns nil at once instead of running out its delay.
// It returns the number of waiters released.
func (wmh *WaitResponseMessageHandler) CancelConnection(connectionId string) int {
	wmh.lock.Lock()
	defer wmh.lock.Unlock()

	released := 0
	for messageId, w := range wmh.waiters {
		if w.connectionId == connectionId {
			// closed and deleted under one lock hold, like everywhere else,
			// so nobody can close or send to this channel afterwards
			close(w.mChannel)
			delete(wmh.waiters, messageId)
			released++
		}
	}
	return released
}

func (wmh *WaitResponseMessageHandler) Handle(connectionId string, messageId string, payload message_common.MessageInterface) {
	if messageId == "" {
		return
	}
	wmh.lock.Lock()
	defer wmh.lock.Unlock()

	waiter, ok := wmh.waiters[messageId]

	if !ok {
		logger.Warn("No waiter for messageId", messageId)
		return
	}

	if waiter.connectionId != connectionId {
		logger.Warnf("the waiter's connection id %s is not the same as received %s by message id %s", waiter.connectionId, connectionId, messageId)
		close(waiter.mChannel)
		delete(wmh.waiters, messageId)
		return
	}

	select {
	case waiter.mChannel <- payload:
	default:
		logger.Warnf("waiter mChannel for messageId %s was already full, dropping message", messageId)
	}
	delete(wmh.waiters, messageId)
}

// the meat

func (wmh *WaitResponseMessageHandler) WaitForResponse(connectionId string, messageId string, delay time.Duration) message_common.MessageInterface {
	if messageId == "" {
		return nil
	}

	waiter := &Waiter{
		connectionId: connectionId,
		messageId:    messageId,
		mChannel:     make(chan message_common.MessageInterface, 1),
	}

	if !wmh.AddWaiter(waiter) {
		logger.Warnf("There is another waiter for a message id %s", messageId)
		close(waiter.mChannel)
		return nil
	}
	// backstop: guarantees cleanup on every exit (no-op if Handle already consumed it)
	defer wmh.RemoveWaiter(waiter)

	select {
	case message, ok := <-waiter.mChannel:
		if !ok {
			logger.Warnf("waiter channel for messageId %s was closed without a response", messageId)
			return nil
		}
		return message
	case <-time.After(delay):
		return nil
	}
}
