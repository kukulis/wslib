package handler

import (
	"sync"
	"testing"
	"time"

	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/pkg/message_common"
)

func TestWaitForResponse_HappyPath(t *testing.T) {
	wmh := NewWaitResponseMessageHandler()
	payload := messages.NewDefaultMessage(messages.MessageTypeDefault, "resp-1")

	done := make(chan message_common.MessageInterface, 1)
	go func() {
		done <- wmh.WaitForResponse("client-1", "msg-1", time.Second)
	}()

	// give the goroutine a moment to register before responding
	time.Sleep(10 * time.Millisecond)
	wmh.Handle("client-1", "msg-1", payload)

	select {
	case got := <-done:
		if got != payload {
			t.Errorf("expected returned message to be the same payload, got %v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForResponse did not return in time")
	}
}

func TestWaitForResponse_TimeoutReturnsNil(t *testing.T) {
	wmh := NewWaitResponseMessageHandler()

	got := wmh.WaitForResponse("client-1", "msg-timeout", 20*time.Millisecond)
	if got != nil {
		t.Errorf("expected nil on timeout, got %v", got)
	}

	wmh.lock.Lock()
	_, ok := wmh.waiters["msg-timeout"]
	wmh.lock.Unlock()
	if ok {
		t.Error("expected waiter to be removed from map after timeout")
	}
}

func TestWaitForResponse_EmptyMessageIdReturnsNilImmediately(t *testing.T) {
	wmh := NewWaitResponseMessageHandler()

	got := wmh.WaitForResponse("client-1", "", time.Second)
	if got != nil {
		t.Errorf("expected nil for empty messageId, got %v", got)
	}
}

func TestHandle_UnknownMessageIdDoesNotPanic(t *testing.T) {
	wmh := NewWaitResponseMessageHandler()
	payload := messages.NewDefaultMessage(messages.MessageTypeDefault, "resp-1")

	// no waiter registered; must be a no-op, not a panic
	wmh.Handle("client-1", "no-such-id", payload)
}

func TestHandle_EmptyMessageIdIsNoOp(t *testing.T) {
	wmh := NewWaitResponseMessageHandler()
	payload := messages.NewDefaultMessage(messages.MessageTypeDefault, "resp-1")

	wmh.Handle("client-1", "", payload)
}

func TestHandle_ClientIdMismatchClosesChannelAndReturnsNil(t *testing.T) {
	wmh := NewWaitResponseMessageHandler()
	payload := messages.NewDefaultMessage(messages.MessageTypeDefault, "resp-1")

	done := make(chan message_common.MessageInterface, 1)
	go func() {
		done <- wmh.WaitForResponse("client-A", "msg-mismatch", time.Second)
	}()

	time.Sleep(10 * time.Millisecond)
	wmh.Handle("client-B", "msg-mismatch", payload) // wrong clientId

	select {
	case got := <-done:
		if got != nil {
			t.Errorf("expected nil on clientId mismatch, got %v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForResponse did not return in time")
	}

	wmh.lock.Lock()
	_, ok := wmh.waiters["msg-mismatch"]
	wmh.lock.Unlock()
	if ok {
		t.Error("expected waiter to be removed from map after mismatch")
	}
}

func TestWaitForResponse_DuplicateMessageIdSecondCallerRejectedImmediately(t *testing.T) {
	wmh := NewWaitResponseMessageHandler()
	payload := messages.NewDefaultMessage(messages.MessageTypeDefault, "resp-1")

	firstStarted := make(chan struct{})
	firstDone := make(chan message_common.MessageInterface, 1)
	go func() {
		// register first, then signal, then block waiting for a response
		w := &Waiter{clientId: "client-1", messageId: "dup-id", mChannel: make(chan message_common.MessageInterface, 1)}
		if !wmh.AddWaiter(w) {
			t.Error("expected first AddWaiter to succeed")
			close(firstStarted)
			firstDone <- nil
			return
		}
		close(firstStarted)
		select {
		case m := <-w.mChannel:
			firstDone <- m
		case <-time.After(time.Second):
			firstDone <- nil
		}
	}()

	<-firstStarted

	start := time.Now()
	second := wmh.WaitForResponse("client-2", "dup-id", time.Second)
	elapsed := time.Since(start)

	if second != nil {
		t.Errorf("expected second caller to get nil, got %v", second)
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("expected second caller to be rejected immediately, took %v", elapsed)
	}

	wmh.Handle("client-1", "dup-id", payload)

	if got := <-firstDone; got != payload {
		t.Errorf("expected first caller to receive the payload, got %v", got)
	}
}

func TestHandle_SecondSendToFullBufferDoesNotBlockOrPanic(t *testing.T) {
	wmh := NewWaitResponseMessageHandler()
	first := messages.NewDefaultMessage(messages.MessageTypeDefault, "first")
	second := messages.NewDefaultMessage(messages.MessageTypeDefault, "second")

	w := &Waiter{clientId: "client-1", messageId: "full-buffer", mChannel: make(chan message_common.MessageInterface, 1)}
	if !wmh.AddWaiter(w) {
		t.Fatal("expected AddWaiter to succeed")
	}

	// simulate a prior send already having filled the buffer
	w.mChannel <- first

	finished := make(chan struct{})
	go func() {
		wmh.Handle("client-1", "full-buffer", second) // must hit the default branch, not block
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("Handle blocked instead of taking the default branch on a full buffer")
	}

	got := <-w.mChannel
	if got != first {
		t.Errorf("expected the original buffered message to remain, got %v", got)
	}
}

func TestWaitForResponse_ConcurrentDuplicateCallersNoPanicNoDeadlock(t *testing.T) {
	wmh := NewWaitResponseMessageHandler()
	payload := messages.NewDefaultMessage(messages.MessageTypeDefault, "resp-1")

	const n = 10
	var wg sync.WaitGroup
	results := make([]message_common.MessageInterface, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = wmh.WaitForResponse("client-1", "race-id", 300*time.Millisecond)
		}(i)
	}

	// give everyone a chance to register/race, then respond once
	time.Sleep(20 * time.Millisecond)
	wmh.Handle("client-1", "race-id", payload)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("goroutines did not finish: suspected deadlock")
	}

	gotPayload := 0
	for _, r := range results {
		if r == payload {
			gotPayload++
		}
	}
	if gotPayload > 1 {
		t.Errorf("expected at most one caller to receive the payload, got %d", gotPayload)
	}
}
