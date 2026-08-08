package tcsync

import (
	"context"
	"errors"
	"sync"

	"go.llib.dev/testcase/internal/errorkitlite"
	"go.llib.dev/testcase/sandbox"
)

type Job interface {
	Wait() error
	Done() <-chan struct{}
	Cancel()
}

type Group struct {
	// Isolation ensures that each function in a Group runs separately,
	// so if one encounters an error, it won’t affect the others.
	Isolation bool
	// ErrorOnGoexit makes Group#Wait and the Job#Wait of the jobs created within the group to return with error
	// if goroutine was stopped due to `runtime.Goexit()`.
	ErrorOnGoexit bool

	wg sync.WaitGroup

	rwm     sync.RWMutex
	_done   chan struct{}
	cancels map[int]func()
	errs    []error
	panic   *any
	subs    map[int]*Group
}

func (g *Group) Len() int {
	g.rwm.RLock()
	defer g.rwm.RUnlock()
	return g.len()
}

func (g *Group) len() int {
	return len(g.cancels)
}

type ErrGoexit struct{}

func (ErrGoexit) Error() string { return "ErrGoexit" }

// Go calls fn in a new goroutine and adds that task to the [Group].
// When fn returns, the task is removed from the Group.
//
// If the Group is empty, Go must happen before a [Group.Wait].
// Typically, this simply means Go is called to start tasks before Wait is called.
// If the Group is not empty, Go may happen at any time.
// This means a goroutine started by Go may itself call Go.
// If a Group is reused to wait for several independent sets of tasks,
// new Go calls must happen after all previous Wait calls have returned.
//
// In the terminology of [the Go memory model], the return from fn
// "synchronizes before" the return of any Wait call that it unblocks.
//
// [the Go memory model]: https://go.dev/ref/mem
func (g *Group) Go(ctx context.Context, fn func(ctx context.Context) error) Job {
	if ctx == nil {
		ctx = context.Background()
	}

	g.rwm.Lock()
	defer g.rwm.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	if g.cancels == nil {
		g.cancels = make(map[int]func())
	}

	id := nextID(g.cancels)
	g.cancels[id] = cancel

	g.wg.Add(1)

	var (
		m    sync.Mutex
		done = make(chan struct{})
		rerr error
	)
	go func() {
		defer close(done)
		defer g.sigDone()
		defer g.wg.Done()
		var err error
		o := sandbox.Run(func() {
			err = fn(ctx)
		})
		if gerr := g.filterErr(ctx, err); gerr != nil && !g.Isolation {
			g.Cancel()
		}
		m.Lock()
		rerr = err
		m.Unlock()
		g.rwm.Lock()
		defer g.rwm.Unlock()
		if err != nil {
			g.errs = append(g.errs, err)
		}
		if o.Panic {
			g.panic = &o.PanicValue
		}
		if g.ErrorOnGoexit && o.Goexit {
			g.errs = append(g.errs, ErrGoexit{})
		}
		delete(g.cancels, id)
	}()
	j := &job{
		context: ctx,
		cancel:  cancel,
		done:    done,
		err:     &rerr,
	}
	return j
}

func (g *Group) done() chan struct{} {
	g.rwm.RLock()
	if g._done != nil {
		defer g.rwm.RUnlock()
		return g._done
	}
	g.rwm.RUnlock()
	g.rwm.Lock()
	defer g.rwm.Unlock()
	if g._done != nil {
		return g._done
	}
	g._done = make(chan struct{})
	return g._done
}

func (g *Group) Done() <-chan struct{} {
	g.done() // init done chan
	g.rwm.RLock()
	defer g.rwm.RUnlock()
	if g.len() == 0 {
		d := make(chan struct{})
		close(d)
		return d
	}
	return g._done
}

func (g *Group) sigDone() {
	done := g.done()
	g.rwm.RLock()
	defer g.rwm.RUnlock()
	if g.len() != 0 {
		return
	}
	for {
		select {
		case done <- struct{}{}:
		default:
			return
		}
	}
}

func (g *Group) filterErr(ctx context.Context, err error) error {
	if err == nil {
		return err
	}
	if ctx == nil {
		return err
	}
	if ctxErr := ctx.Err(); ctxErr != nil && errors.Is(err, ctxErr) {
		return nil
	}
	return err
}

func (g *Group) Cancel() {
	g.rwm.Lock()
	defer g.rwm.Unlock()
	for _, cancel := range g.cancels {
		cancel()
	}
	g.cancels = nil
}

func (g *Group) Wait() (rErr error) {
	g.wg.Wait()
	g.rwm.RLock()
	// fast path
	if len(g.errs) == 0 && g.panic == nil {
		defer g.rwm.RUnlock()
		return nil
	}
	g.rwm.RUnlock()
	// slow path
	g.rwm.Lock()
	defer g.rwm.Unlock()
	if g.panic != nil {
		pv := *g.panic
		g.panic = nil
		panic(pv)
	}
	if len(g.errs) != 0 {
		err := errorkitlite.Merge(g.errs...)
		g.errs = nil
		return err
	}
	return nil
}

func nextID[M ~map[int]V, V any](m M) int {
	if m == nil {
		return 0 // zero
	}
	for i := 0; ; i++ {
		if _, ok := m[i]; !ok {
			return i
		}
	}
}

var _ Job = (*job)(nil)

type job struct {
	context context.Context
	cancel  func()
	done    chan struct{}
	err     *error
}

func (j *job) Done() <-chan struct{} {
	return j.done
}

func (j *job) Wait() error {
	<-j.done
	return *j.err
}

func (j *job) Cancel() {
	if j.cancel != nil {
		j.cancel()
	}
}
