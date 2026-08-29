package synckit_test

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/synckit"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/tcsync"
)

func ExampleGroup() {
	var g synckit.Group

	g.Go(nil, func(ctx context.Context) error {

		return nil
	})

	g.Wait()
}

func TestGroup(t *testing.T) {
	s := testcase.NewSpec(t)

	const timeout = time.Second / 8

	group := let.Var(s, func(t *testcase.T) *synckit.Group {
		return &synckit.Group{}
	})

	var SpecGo = func(s *testcase.Spec, method func(g *synckit.Group, fn func(context.Context) error)) {
		var (
			done = let.Var(s, func(t *testcase.T) chan struct{} {
				ch := make(chan struct{})
				t.Cleanup(func() {
					defer func() { _ = recover() }()
					close(ch)
				})
				return ch
			})
			fnErr = let.VarOf[error](s, nil)
			ran   = let.VarOf(s, false)
			fn    = let.Var(s, func(t *testcase.T) func(context.Context) error {
				return func(ctx context.Context) error {
					ran.Set(t, true)
					select {
					case <-done.Get(t): // task is done
					case <-ctx.Done(): // group requested cancellation
					case <-t.Done(): // test is done
					}
					return fnErr.Get(t)
				}
			})
		)
		act := let.Act0(func(t *testcase.T) {
			method(group.Get(t), fn.Get(t))
		})

		s.Then("it will start the function in the background", func(t *testcase.T) {
			assert.Equal(t, 0, group.Get(t).Len())

			assert.Within(t, time.Millisecond, func(ctx context.Context) {
				act(t)
			})

			assert.Equal(t, 1, group.Get(t).Len())

			close(done.Get(t))

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, 0, group.Get(t).Len())
			})
		})

		s.Then("the background function can be cancelled through cancelling the Group", func(t *testcase.T) {
			assert.Within(t, time.Millisecond, func(ctx context.Context) {
				act(t)
			})

			assert.Within(t, time.Millisecond, func(ctx context.Context) {
				t.Random.Repeat(1, 3, func() {
					group.Get(t).Cancel()
				})
			})

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, 0, group.Get(t).Len())
			})

			assert.Within(t, time.Millisecond, func(ctx context.Context) {
				assert.NoError(t, group.Get(t).Wait())
			})
		})

		s.When("the function encounters an error", func(s *testcase.Spec) {
			fnErr.Let(s, let.Error(s).Get)
			s.Before(func(t *testcase.T) { close(done.Get(t)) }) // no blocking on function execution

			s.Then("we get back the error during Wait", func(t *testcase.T) {
				act(t)

				assert.ErrorIs(t, fnErr.Get(t), group.Get(t).Wait())
			})
		})

		s.When("the function panics", func(s *testcase.Spec) {
			expErr := let.Error(s)

			fn.Let(s, func(t *testcase.T) func(context.Context) error {
				return func(ctx context.Context) error {
					panic(expErr.Get(t))
				}
			})

			s.Then("then on wait, we get the panic back", func(t *testcase.T) {
				act(t)

				got := assert.Panic(t, func() {
					group.Get(t).Wait()
				})

				assert.Equal[any](t, got, expErr.Get(t))
			})
		})

		s.When("the function runtime.Goexit", func(s *testcase.Spec) {
			fn.Let(s, func(t *testcase.T) func(context.Context) error {
				return func(ctx context.Context) error {
					runtime.Goexit()
					return nil
				}
			})

			s.Then("then on wait, we get no error, or panic as goexit is not an error issue", func(t *testcase.T) {
				act(t)

				assert.NotPanic(t, func() {
					group.Get(t).Wait()
				})
			})

			s.And("Group#ErrorOnGoexit set to ", func(s *testcase.Spec) {
				group.Let(s, func(t *testcase.T) *synckit.Group {
					g := group.Super(t)
					g.ErrorOnGoexit = true
					return g
				})

				s.Then("then on wait, we get back an error due to the runtime.Goexit call", func(t *testcase.T) {
					act(t)

					assert.ErrorIs(t, group.Get(t).Wait(), synckit.ErrGoexit{})
				})
			})
		})

		s.When("the function returns with context error upon context cancellation", func(s *testcase.Spec) {
			fn.Let(s, func(t *testcase.T) func(context.Context) error {
				return func(ctx context.Context) error {
					select {
					case <-ctx.Done():
					case <-t.Done():
					}
					return ctx.Err()
				}
			})

			s.Then("upon cancellation, the information that the task was cancelled leaks back with #Wait", func(t *testcase.T) {
				act(t)

				group.Get(t).Cancel()

				assert.ErrorIs(t, group.Get(t).Wait(), context.Canceled)
			})
		})

		s.When("a nested Go call is made within the initial Go call", func(s *testcase.Spec) {
			p1 := let.Phaser(s)
			p2 := let.Phaser(s)

			fn.Let(s, func(t *testcase.T) func(context.Context) error {
				return func(ctx context.Context) error {
					p1.Get(t).Wait()

					group.Get(t).Go(t.Context(), func(ctx context.Context) error {
						p2.Get(t).Wait()
						return nil
					})
					return nil
				}
			})

			s.Then("a #Wait that started to wait initially on the first Go call will also wait for the nested Go", func(t *testcase.T) {
				act(t)

				t.Log("given the main Go call already got CPU time")
				t.Eventually(func(t *testcase.T) {
					assert.Equal(t, p1.Get(t).Len(), 1)
				})

				t.Log("and we start to wait on the group")
				w1 := assert.NotWithin(t, timeout, func(ctx context.Context) {
					group.Get(t).Wait()
				})

				t.Log("and then the nested Go call is started")
				p1.Get(t).Finish()
				t.Eventually(func(t *testcase.T) {
					assert.Equal(t, p2.Get(t).Len(), 1)
				})

				t.Log("then the wait will remain waiting due to the nested Go call")
				w2 := assert.NotWithin(t, timeout, func(ctx context.Context) {
					w1.Wait()
				})

				t.Log("but when the nested Go call finish up too")
				p2.Get(t).Finish()

				t.Log("then the group Wait finally finishes up")
				assert.Within(t, timeout, func(ctx context.Context) {
					w2.Wait()
				})
			})
		})
		s.When("a function is already running in the background", func(s *testcase.Spec) {
			othCancelled := let.VarOf(s, false)

			s.Before(func(t *testcase.T) {
				group.Get(t).Go(nil, func(ctx context.Context) error {
					select {
					case <-t.Done():
					case <-ctx.Done():
						othCancelled.Set(t, true)
					}
					return ctx.Err()
				})
				t.Eventually(func(t *testcase.T) {
					assert.Equal(t, 1, group.Get(t).Len())
				})
			})

			s.Then("it starts a new background task", func(t *testcase.T) {
				assert.Within(t, time.Millisecond, func(ctx context.Context) {
					act(t)
				})

				assert.Equal(t, 2, group.Get(t).Len())
			})

			s.And("if the function encounters an error", func(s *testcase.Spec) {
				fnErr.Let(s, let.Error(s).Get)

				s.Before(func(t *testcase.T) {
					close(done.Get(t)) // no blocking on function execution
				})

				s.Then("it will cancel the other goroutine's context", func(t *testcase.T) {
					act(t)

					t.Eventually(func(t *testcase.T) {
						assert.True(t, othCancelled.Get(t))
					})
				})

				s.And("Isolation was set to true", func(s *testcase.Spec) {
					group.Let(s, func(t *testcase.T) *synckit.Group {
						g := group.Super(t)
						g.Isolation = true
						return g
					})

					s.Then("it will NOT affect the other goroutines", func(t *testcase.T) {
						act(t)

						t.Random.Repeat(42, 128, func() {
							runtime.Gosched()

							assert.False(t, othCancelled.Get(t))
						})
					})
				})
			})
		})
	}

	s.Describe("#Go", func(s *testcase.Spec) {
		var (
			Context, Cancel = let.ContextWithCancel(s)

			isReady    = let.VarOf(s, false)
			isFinished = let.VarOf(s, false)
			isCtxDone  = let.VarOf(s, false)

			fn = let.Var(s, func(t *testcase.T) func(context.Context) error {
				return func(ctx context.Context) error {
					isReady.Set(t, true)
					defer isFinished.Set(t, true)
					select {
					case <-ctx.Done():
						isCtxDone.Set(t, true)
					case <-t.Done():
					}
					return ctx.Err()
				}
			})
		)
		act := let.Act0(func(t *testcase.T) {
			group.Get(t).Go(Context.Get(t), fn.Get(t))
		})

		SpecGo(s, func(g *synckit.Group, fn func(context.Context) error) {
			g.Go(context.Background(), fn)
		})

		s.When("we start a goroutine within the group", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				assert.Within(t, time.Millisecond, func(ctx context.Context) {
					act(t)
				})
				t.Eventually(func(t *testcase.T) {
					assert.True(t, isReady.Get(t))
				})
				t.Random.Repeat(42, 128, func() {
					runtime.Gosched()
					assert.False(t, isFinished.Get(t))
					assert.False(t, isCtxDone.Get(t))
				})
			})

			s.And("the input context of this goroutine is cancelled", func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					Cancel.Get(t)()
				})

				s.Then("the started goroutine's context will be cancelled", func(t *testcase.T) {
					t.Eventually(func(t *testcase.T) {
						assert.True(t, isCtxDone.Get(t))
					})
					t.Eventually(func(t *testcase.T) {
						assert.True(t, isFinished.Get(t))
					})
				})

				s.And("other goroutine in the group was already started previously", func(s *testcase.Spec) {
					var (
						isOthReady    = let.VarOf(s, false)
						isOthFinished = let.VarOf(s, false)
						isOthCtxDone  = let.VarOf(s, false)
					)
					group.Let(s, func(t *testcase.T) *synckit.Group {
						g := group.Super(t)
						g.Go(context.Background(), func(ctx context.Context) error {
							isOthReady.Set(t, true)
							defer isOthFinished.Set(t, true)
							select {
							case <-ctx.Done():
								isOthCtxDone.Set(t, true)
							case <-t.Done():
							}
							return ctx.Err()
						})
						return g
					})

					s.Then("other process will not be affected by the current execution context's cancellation", func(t *testcase.T) {
						t.Eventually(func(t *testcase.T) {
							assert.True(t, isCtxDone.Get(t))
						})
						t.Eventually(func(t *testcase.T) {
							assert.True(t, isOthReady.Get(t))
						})
						t.Random.Repeat(42, 128, func() {
							runtime.Gosched()
							assert.False(t, isOthCtxDone.Get(t))
						})
					})
				})
			})
		})
	})

	s.Context("#Cancel", func(s *testcase.Spec) {
		s.Test("Group", func(t *testcase.T) {
			var (
				g synckit.Group
				n atomic.Int32
				p tcsync.Phaser
			)
			defer p.Finish()

			g.Go(t.Context(), func(ctx context.Context) error {
				defer n.Add(1)
				<-ctx.Done()
				return nil
			})

			g.Go(t.Context(), func(ctx context.Context) error {
				defer n.Add(1)
				<-ctx.Done()
				return nil
			})

			g.Go(t.Context(), func(ctx context.Context) error {
				defer n.Add(1)
				<-ctx.Done()
				return nil
			})

			g.Cancel()

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, n.Load(), 3)
			})
		})

		s.Test("GroupJob", func(t *testcase.T) {
			var (
				g synckit.Group
				n atomic.Int32
				p tcsync.Phaser
			)
			defer p.Finish()

			j1 := g.Go(t.Context(), func(ctx context.Context) error {
				defer n.Add(1)
				<-ctx.Done()
				return nil
			})

			j2 := g.Go(t.Context(), func(ctx context.Context) error {
				defer n.Add(1)
				<-ctx.Done()
				return nil
			})

			g.Go(t.Context(), func(ctx context.Context) error {
				defer n.Add(1)
				<-ctx.Done()
				return nil
			})

			j1.Cancel()

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, n.Load(), 1)
			})

			t.Random.Repeat(3, 7, func() {
				runtime.Gosched()

				assert.Equal(t, n.Load(), 1)
			})

			j2.Cancel()

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, n.Load(), 2)
			})
			t.Random.Repeat(3, 7, func() {
				runtime.Gosched()

				assert.Equal(t, n.Load(), 2)
			})

			g.Cancel()

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, n.Load(), 3)
			})
		})

		s.Test("a task started after a Cancel is still cancellable once the earlier tasks finished", func(t *testcase.T) {
			var g synckit.Group

			var p tcsync.Phaser
			defer p.Finish()

			// outlives the Cancel, as it stops on the phaser, not on its context
			g.Go(t.Context(), func(ctx context.Context) error {
				p.Wait()
				return nil
			})

			g.Cancel()

			g.Go(t.Context(), func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			})

			// let the first task finish, so that it cleans up after itself
			p.Finish()

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, g.Len(), 1,
					"expected that only the task started after the Cancel is left")
			})

			g.Cancel()

			assert.Within(t, timeout, func(ctx context.Context) {
				assert.NoError(t, g.Wait())
			}, "expected that a task started after a previous Cancel is still cancellable",
				"a finishing task must not take another task's cancellation with it")
		})
	})

	s.Describe("#Done", func(s *testcase.Spec) {
		s.Test("on empty group Done acts akin to Group#wait and reports that the group is done", func(t *testcase.T) {
			var g synckit.Group

			assert.Within(t, timeout, func(ctx context.Context) {
				select {
				case <-g.Done():
				case <-t.Done():
				}
			}, "expect that empty group act as being done already")
		})

		s.Test("Done will wait until goroutines are finished", func(t *testcase.T) {
			var g synckit.Group
			sampling := t.Random.IntBetween(3, 7)

			var p tcsync.Phaser
			defer p.Finish()

			for i := 0; i < sampling; i++ {
				g.Go(t.Context(), func(ctx context.Context) error {
					p.Wait()
					return nil
				})
			}

			waitingOnDone := assert.NotWithin(t, timeout, func(ctx context.Context) {
				select {
				case <-g.Done():
				case <-t.Done():
				}
			})

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, p.Len(), sampling,
					"expect that all the go routine are waiting on the phaser")
			})

			releaseCount := t.Random.IntN(sampling) // apart from one, other go routines are released
			assert.NotEqual(t, 0, sampling-releaseCount,
				"test implementation bug, it was not expected that IntN(sampling) returns with the count of sampling")

			curr := sampling
			for i := 0; i < releaseCount; i++ {
				p.Signal()
				curr--
				t.Eventually(func(t *testcase.T) {
					assert.Equal(t, p.Len(), curr)
					assert.Equal(t, g.Len(), curr)
				})
			}

			p.Finish()

			assert.Within(t, timeout, func(ctx context.Context) {
				waitingOnDone.Wait()
			}, "expected that all waiting processes finished waiting",
				"including the ones started after goroutines were already fired")

		})

		s.Test("Done keeps blocking while a cancelled goroutine is still shutting down", func(t *testcase.T) {
			var g synckit.Group

			var p tcsync.Phaser
			defer p.Finish()

			g.Go(t.Context(), func(ctx context.Context) error {
				<-ctx.Done()
				// Cancel only asked the goroutine to stop,
				// it is still running until its shutdown is over.
				p.Wait()
				return nil
			})

			assert.Equal(t, g.Len(), 1)

			g.Cancel()

			t.Eventually(func(t *testcase.T) {
				assert.Equal(t, p.Len(), 1,
					"expected that the cancelled goroutine began its shutdown")
			})

			assert.Equal(t, g.Len(), 1,
				"expected that cancelling the group doesn't make a still running goroutine disappear")

			waitingOnDone := assert.NotWithin(t, timeout, func(ctx context.Context) {
				select {
				case <-g.Done():
				case <-t.Done():
				}
			}, "expected that Done keeps blocking while a cancelled goroutine is still finishing")

			p.Finish()

			assert.Within(t, timeout, func(ctx context.Context) {
				waitingOnDone.Wait()
			}, "expected that Done is released once the goroutine actually finished")
		})

		s.Test("GroupJob", func(t *testcase.T) {
			var p tcsync.Phaser
			defer p.Finish()

			var g synckit.Group

			job := g.Go(t.Context(), func(ctx context.Context) error {
				p.Wait()
				return nil
			})

			w := assert.NotWithin(t, timeout, func(ctx context.Context) {
				select {
				case <-job.Done():
				case <-t.Done():
				}
			}, "expect that job.Done blocks while the job is ")

			p.Finish()

			assert.Within(t, timeout, func(ctx context.Context) {
				w.Wait()
			}, "expect that job finish, the job.Done is finished")

			assert.Within(t, timeout, func(ctx context.Context) {
				select {
				case <-job.Done():
				case <-t.Done():
				}
			}, "expect that a job that already finished no longer blocks on Done()")
		})

		s.Test("race", func(t *testcase.T) {
			var g synckit.Group
			var p tcsync.Phaser

			// race when Done is after Go
			testcase.Race(func() {
				g.Go(t.Context(), func(ctx context.Context) error {
					p.Wait()
					return nil
				})
			}, func() {
				t.Eventually(func(t *testcase.T) {
					assert.Equal(t, 1, p.Len())
				})
				p.Broadcast()
				<-g.Done()
			})

			// race when Go is after Done (empty goroutine)
			testcase.Race(func() {
				wait()
				g.Go(t.Context(), func(ctx context.Context) error {
					return nil
				})
			}, func() {
				<-g.Done()
			})
		})
	})

	s.Describe("#Wait", func(s *testcase.Spec) {
		var (
			group = let.Var(s, func(t *testcase.T) *synckit.Group {
				return &synckit.Group{}
			})
			fnErr = let.Var(s, func(t *testcase.T) error {
				return fmt.Errorf("the job failed: %s", t.Random.UUID())
			}).EagerLoading(s)
			job = let.Var(s, func(t *testcase.T) synckit.Job {
				err := fnErr.Get(t)
				return group.Get(t).Go(t.Context(), func(ctx context.Context) error {
					return err
				})
			}).EagerLoading(s)
		)

		act := let.Act(func(t *testcase.T) error {
			return group.Get(t).Wait()
		})

		s.Then("it returns the error which the job failed with", func(t *testcase.T) {
			assert.ErrorIs(t, act(t), fnErr.Get(t))
		})

		s.Then("it reports the error only once", func(t *testcase.T) {
			assert.ErrorIs(t, act(t), fnErr.Get(t))
			assert.NoError(t, act(t))
		})

		s.When("the job's error was taken over with Job#Wait", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				assert.ErrorIs(t, job.Get(t).Wait(), fnErr.Get(t))
			})

			s.Then("it doesn't report the error again", func(t *testcase.T) {
				assert.NoError(t, act(t),
					"expected that whoever waited on the job is trusted to deal with the error they took over")
			})

			s.And("another job failed which no one waited on", func(s *testcase.Spec) {
				othErr := let.Var(s, func(t *testcase.T) error {
					return fmt.Errorf("the other job failed: %s", t.Random.UUID())
				}).EagerLoading(s)

				s.Before(func(t *testcase.T) {
					err := othErr.Get(t)
					group.Get(t).Go(t.Context(), func(ctx context.Context) error {
						return err
					})
				})

				s.Then("it reports that job's error", func(t *testcase.T) {
					got := act(t)

					assert.ErrorIs(t, got, othErr.Get(t))
					assert.NotContains(t, got.Error(), fnErr.Get(t).Error(),
						"expected that taking over one job's error doesn't hide another job's")
				})
			})
		})

		s.When("the job was only observed with Job#Done", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				assert.Within(t, timeout, func(ctx context.Context) {
					select {
					case <-job.Get(t).Done():
					case <-t.Done():
					}
				})
			})

			s.Then("it still reports the error", func(t *testcase.T) {
				assert.ErrorIs(t, act(t), fnErr.Get(t),
					"expected that observing that a job is over is not the same as dealing with its error")
			})
		})
	})

	s.Test("race", func(t *testcase.T) {
		var g synckit.Group
		defer g.Wait()

		const sampling = 42
		var work = func(ctx context.Context) error {
			for i := 0; i < sampling; i++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-t.Done():
					return nil
				default:
					runtime.Gosched()
				}
			}
			return nil
		}

		j := g.Go(t.Context(), work)

		testcase.Race(func() {
			wait()
			g.Go(t.Context(), work)
		}, func() {
			wait()
			for i := 0; i < sampling; i++ {
				g.Len()
			}
		}, func() {
			<-g.Done()
		}, func() {
			g.Wait()
		}, func() {
			wait()
			g.Cancel()
		}, func() {
			j.Wait()
		})
	})
}

func BenchmarkGroup(b *testing.B) {
	wrk := func() {
		for i := 0; i < 128; i++ {
			_ = 5 * i
		}
	}

	b.Run("with boilerplate", func(b *testing.B) {
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				wrk()
			}()
			wg.Wait()
		}
	})

	b.Run("group", func(b *testing.B) {
		var g synckit.Group
		for i := 0; i < b.N; i++ {
			g.Go(nil, func(ctx context.Context) error {
				return nil
			})
			g.Wait()
		}
	})
}

func wait() {
	for i := 0; i < 128; i++ {
		runtime.Gosched()
		time.Sleep(time.Nanosecond)
	}
}
