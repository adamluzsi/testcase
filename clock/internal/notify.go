package internal

import (
	"sync"
	"time"
)

// handler is a registered time travel event subscriber.
type handler struct {
	channel chan<- TimeTravelEvent
	// done is closed when the handler is unregistered.
	// It releases publish goroutines still blocked on the send,
	// after the receiver goroutine has already exited.
	done chan struct{}
}

var handlers = make(map[int]handler)

type TimeTravelEvent struct {
	Deep   bool
	Freeze bool
	When   time.Time
	Prev   time.Time
}

func Notify(c chan<- TimeTravelEvent) func() {
	if c == nil {
		panic("clock: Notify using nil channel")
	}
	defer mSync()()
	var index int
	for i := 0; true; i++ {
		if _, ok := handlers[i]; !ok {
			index = i
			break
		}
	}
	h := handler{channel: c, done: make(chan struct{})}
	handlers[index] = h
	var once sync.Once
	return func() {
		defer mSync()()
		delete(handlers, index)
		// release any publish goroutine still waiting to deliver to c.
		once.Do(func() { close(h.done) })
	}
}

func Check() (TimeTravelEvent, bool) {
	defer mRSync()()
	return lookupTimeTravelEvent()
}

func lookupTimeTravelEvent() (TimeTravelEvent, bool) {
	return TimeTravelEvent{
		Deep:   chrono.Timeline.Deep,
		Freeze: chrono.Timeline.Frozen,
		When:   chrono.Timeline.When,
		Prev:   chrono.Timeline.Prev,
	}, !chrono.Timeline.IsZero()
}

func notify() {
	defer mRSync()()
	tt, _ := lookupTimeTravelEvent()
	var publish = func(h handler) {
		defer recover()
		select {
		case h.channel <- tt:
		case <-h.done:
		}
	}
	for _, h := range handlers {
		go publish(h)
	}
}
