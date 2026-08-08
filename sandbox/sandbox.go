package sandbox

import (
	"runtime"
)

func Run(fn func()) O {
	var done = make(chan struct{})
	var o O
	go func() {
		defer close(done)

		var r any
		defer func() {
			if o.OK {
				return
			}
			if o.Panic {
				if r == nil {
					return
				}
				if _, isNilPanic := r.(*runtime.PanicNilError); isNilPanic {
					return
				}
				o.PanicValue = r
			} else {
				o.Goexit = true
			}
		}()
		func() {
			defer func() {
				if !o.OK {
					o.Frames = getFrames()
				}
				r = recover()
			}()
			fn()
			o.OK = true
		}()
		if !o.OK { // panic:true, goexit:false
			o.Panic = true
		}
	}()
	<-done
	return o
}

type O struct {
	OK     bool
	Goexit bool
	Panic  bool

	PanicValue any
	Frames     []runtime.Frame
}
