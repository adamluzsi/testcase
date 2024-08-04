package internal

import (
	"time"
)

var handlers = make(map[int]chan<- TimeTravelEvent)

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
	defer lock()()
	var index int
	for i := 0; true; i++ {
		if _, ok := handlers[i]; !ok {
			index = i
			break
		}
	}
	handlers[index] = c
	return func() {
		defer lock()()
		delete(handlers, index)
	}
}

func Check() (TimeTravelEvent, bool) {
	defer rlock()()
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
	defer rlock()()
	tt, _ := lookupTimeTravelEvent()
	var publish = func(channel chan<- TimeTravelEvent) {
		defer recover()
		channel <- tt
	}
	for _, ch := range handlers {
		go publish(ch)
	}
}
