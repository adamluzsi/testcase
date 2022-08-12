package internal

func init() {
	notify()
}

var change chan struct{}

func Listen() <-chan struct{} {
	defer rlock()()
	return change
}

func notify() {
	defer lock()()
	if change != nil {
		close(change)
	}
	change = make(chan struct{})
}
