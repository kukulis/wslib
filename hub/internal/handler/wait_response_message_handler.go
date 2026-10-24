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

	delete(wmh.waiters, w.messageId)
}

func (wmh *WaitResponseMessageHandler) Handle(clientId string, messageId string, payload message_common.MessageInterface) {
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

	if waiter.clientId != clientId {
		logger.Warnf("the waiter's client id %s is not the same as received %s by message id %s", waiter.clientId, clientId, messageId)
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

func (wmh *WaitResponseMessageHandler) WaitForResponse(clientId string, messageId string, delay time.Duration) message_common.MessageInterface {
	if messageId == "" {
		return nil
	}

	waiter := &Waiter{
		clientId:  clientId,
		messageId: messageId,
		mChannel:  make(chan message_common.MessageInterface, 1),
	}

	if !wmh.AddWaiter(waiter) {
		logger.Warnf("There is another waiter for a message id %s", messageId)
		close(waiter.mChannel)
		return nil
	}

	select {
	case message, ok := <-waiter.mChannel:
		if !ok {
			logger.Warnf("waiter channel for messageId %s was closed without a response", messageId)
			return nil
		}
		return message
	case <-time.After(delay):
		wmh.RemoveWaiter(waiter)
		return nil
	}
}
