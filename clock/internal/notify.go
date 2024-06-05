package internal

var handlers = map[int]chan<- struct{}{}

func Notify(c chan<- struct{}) func() {
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

func notify() {
	defer rlock()()
	for index, ch := range handlers {
		go func(i int, ch chan<- struct{}) {
			defer recover()
			ch <- struct{}{}
		}(index, ch)
	}
}
