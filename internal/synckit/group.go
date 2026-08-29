package synckit

import (
	"context"
	"errors"
	"sync"

	"go.llib.dev/testcase/internal/errorkitlite"
	"go.llib.dev/testcase/sandbox"
)

type Job interface {
	// Wait blocks until the job is finished, then returns its error.
	//
	// Waiting on a job takes its error over: the caller is expected to deal
	// with it, so [Group.Wait] no longer reports it.
	Wait() error
	// Done is closed when the job is finished.
	//
	// Unlike Wait, observing that a job is over doesn't take over its error.
	Done() <-chan struct{}
	// Cancel cancels the job's context, asking it to stop.
	Cancel()
}

type Group struct {
	// Isolation ensures that each function in a Group runs separately,
	// so if one encounters an error, it won’t affect the others.
	Isolation bool
	// ErrorOnGoexit makes the Group report an [ErrGoexit] for a task that was stopped by `runtime.Goexit()`.
	//
	// Such a task never returned, so it had no chance to tell what went wrong,
	// and [ErrGoexit] takes the place of the error it would have returned.
	ErrorOnGoexit bool

	wg sync.WaitGroup

	rwm     sync.RWMutex
	_done   chan struct{}
	running int
	cancels map[int]func()
	jobs    []*job
	panic   *any
	subs    map[int]*Group
}

// Len reports how many tasks of the [Group] are currently running.
//
// A task which was asked to stop with [Group.Cancel] is still counted
// until it actually returns, since cancelling only asks it to stop.
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
// The error of fn belongs to the returned [Job].
// Whoever takes it over with [Job.Wait] is expected to deal with it,
// so from that point on [Group.Wait] no longer reports it.
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

	isRunning := 0 < g.len()

	id := nextID(g.cancels)
	g.cancels[id] = cancel

	if !isRunning {
		// the previous batch's done channel is already closed,
		// so this batch needs a fresh one to be waited on
		g._done = nil
	}
	g.running++

	g.wg.Add(1)

	j := &job{
		context: ctx,
		cancel:  cancel,
		done:    make(chan struct{}),
	}
	g.jobs = append(g.jobs, j)

	go func() {
		// registration order is the reverse of the execution order,
		// so the group only reports the task as finished after the task's
		// own completion is fully observable
		defer g.jobDone(id)
		defer g.wg.Done()
		defer close(j.done)
		var err error
		o := sandbox.Run(func() {
			err = fn(ctx)
		})
		if gErr := g.filterErr(ctx, err); gErr != nil && !g.Isolation {
			g.Cancel()
		}
		if g.ErrorOnGoexit && o.Goexit {
			// fn never returned, so it had no chance to tell what went wrong
			err = ErrGoexit{}
		}
		j.setErr(err)
		if o.Panic {
			g.setPanic(o.PanicValue)
		}
	}()
	return j
}

func (g *Group) setPanic(pv any) {
	g.rwm.Lock()
	defer g.rwm.Unlock()
	g.panic = &pv
}

// jobDone removes a finished task from the [Group],
// and releases everyone who waits for the Group to become empty.
func (g *Group) jobDone(id int) {
	g.rwm.Lock()
	defer g.rwm.Unlock()
	delete(g.cancels, id)
	g.running--
	if g.len() == 0 && g._done != nil {
		close(g._done)
	}
}

// done must be called while g.rwm is held for writing.
func (g *Group) done() chan struct{} {
	if g._done == nil {
		g._done = make(chan struct{})
		if g.len() == 0 {
			close(g._done)
		}
	}
	return g._done
}

// Done returns a channel that is closed when the [Group] has no running task.
//
// An empty Group is done already, so the returned channel is closed right away.
// Since [Group.Cancel] only asks the tasks to stop, the channel is not closed
// until they actually returned.
//
// Unlike [Group.Wait], observing that the Group is done doesn't take over the
// errors of the tasks, and it doesn't replay a panic either.
func (g *Group) Done() <-chan struct{} {
	g.rwm.Lock()
	defer g.rwm.Unlock()
	return g.done()
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

// Cancel asks every task of the [Group] to stop by cancelling their context.
//
// It only asks: the tasks keep being part of the Group until they return,
// so [Group.Len] still counts them and [Group.Done] still blocks on them.
func (g *Group) Cancel() {
	g.rwm.Lock()
	defer g.rwm.Unlock()
	for _, cancel := range g.cancels {
		cancel()
	}
}

// Wait blocks until every task of the [Group] is finished,
// then returns the errors which no one else took over.
//
// An error that was already taken over with [Job.Wait] is considered dealt with
// by whoever waited on that job, so it is not reported here.
func (g *Group) Wait() (rErr error) {
	g.wg.Wait()

	g.rwm.Lock()
	defer g.rwm.Unlock()

	if g.panic != nil {
		pv := *g.panic
		g.panic = nil
		panic(pv)
	}

	var (
		errs    []error
		pending []*job
	)
	for _, j := range g.jobs {
		if err := j.takeErr(); err != nil {
			errs = append(errs, err)
		}
		if !j.isDone() {
			// started while we were collecting, so it is not accounted for yet
			pending = append(pending, j)
		}
	}
	g.jobs = pending
	return errorkitlite.Merge(errs...)
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

	mutex sync.Mutex
	err   error
	taken bool
}

func (j *job) Done() <-chan struct{} {
	return j.done
}

// Wait blocks until the job is finished, then returns its error.
//
// Waiting on a job takes its error over: the caller is expected to deal with it,
// so [Group.Wait] no longer reports it.
func (j *job) Wait() error {
	// taken upfront, since the job may only finish while we are waiting here
	j.take()
	<-j.done
	j.mutex.Lock()
	defer j.mutex.Unlock()
	return j.err
}

func (j *job) Cancel() {
	if j.cancel != nil {
		j.cancel()
	}
}

func (j *job) isDone() bool {
	select {
	case <-j.done:
		return true
	default:
		return false
	}
}

func (j *job) setErr(err error) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.err = err
}

func (j *job) take() {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.taken = true
}

// takeErr returns the job's error unless someone already took it over,
// and marks it as taken, so that it is only ever reported once.
func (j *job) takeErr() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	if j.taken {
		return nil
	}
	j.taken = true
	return j.err
}
