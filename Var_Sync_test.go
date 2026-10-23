package testcase_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/tcsync"
)

func TestVar_Sync(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		// a unique key ensures that the concurrently running test processes don't affect the test
		key = let.Var(s, func(t *testcase.T) string {
			return "TestVar_Sync-" + t.Random.UUID()
		})
		// inits counts the initialisations of the Var
		inits = let.Var(s, func(t *testcase.T) *int32 {
			return new(int32)
		})
		// befores counts the executions of the Var's Before hook, when the Var has one
		befores = let.Var(s, func(t *testcase.T) *int32 {
			return new(int32)
		})
		subject = let.Var(s, func(t *testcase.T) testcase.Var[int] {
			n := inits.Get(t)
			return testcase.Var[int]{
				ID: testcase.VarID(key.Get(t)),
				Init: func(t *testcase.T) int {
					atomic.AddInt32(n, 1)
					return t.Random.Int()
				},
				Sync: true,
			}
		})
		// tb is the testing.TB of the test that uses the Var
		tb = let.Var(s, func(t *testcase.T) *testcase.FakeTB {
			return newSyncTB(t, "TestSubject")
		})
		// spec is the Spec where the Var is used
		spec = let.Var(s, func(t *testcase.T) *testcase.Spec {
			return testcase.NewSpec(tb.Get(t))
		})
		// tct is the test of the Spec that uses the Var.
		// Making it starts the test, which runs the hooks of the Spec,
		// and the test lasts until its testing.TB finishes.
		tct = let.Var(s, func(t *testcase.T) *testcase.T {
			return testcase.NewTWithSpec(tb.Get(t), spec.Get(t))
		})
		// holder is another test, which may hold a key
		holder = let.Var(s, func(t *testcase.T) *testcase.FakeTB {
			return newSyncTB(t, "TestHolder")
		})
		holderKey = let.Var(s, func(t *testcase.T) string {
			return key.Get(t)
		})
	)

	s.Describe("#Get", func(s *testcase.Spec) {
		act := func(t *testcase.T) int {
			return subject.Get(t).Get(tct.Get(t))
		}

		s.Then("the value is returned without waiting", func(t *testcase.T) {
			assert.Within(t, syncTimeout, func(context.Context) { act(t) })
		})

		s.Then("other tests wait for the Var's ID as key, until the test finishes", func(t *testcase.T) {
			act(t)
			other := newSyncTB(t, "TestOther")

			w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) {
				testcase.Sync(other, key.Get(t))
			})

			tb.Get(t).Finish()
			assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
		})

		s.When("the Init of the Var registers a cleanup", func(s *testcase.Spec) {
			// otherSync receives the attempt of another test to acquire the key, which is made during the cleanup
			otherSync := let.Var(s, func(t *testcase.T) chan *assert.Async {
				return make(chan *assert.Async, 1)
			})
			subject.Let(s, func(t *testcase.T) testcase.Var[int] {
				v := subject.Super(t)
				init := v.Init
				v.Init = func(tct *testcase.T) int {
					tct.Cleanup(func() {
						other := newSyncTB(t, "TestOther")
						// Should, so the attempt is always sent, even when it fails
						otherSync.Get(t) <- assert.Should(t).NotWithin(syncBlockingTimeout, func(context.Context) {
							testcase.Sync(other, key.Get(t))
						})
					})
					return init(tct)
				}
				return v
			})

			s.Then("the cleanup runs while the test still holds the key", func(t *testcase.T) {
				act(t)

				tb.Get(t).Finish()
				w := <-otherSync.Get(t)
				assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
			})
		})

		s.When("another test holds the key", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testcase.Sync(holder.Get(t), holderKey.Get(t))
			})

			s.Then("it waits until the other test finishes", func(t *testcase.T) {
				w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })

				holder.Get(t).Finish()
				assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
			})

			s.Then("the Var is only initialised once the test holds the key", func(t *testcase.T) {
				w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })
				assert.Equal(t, atomic.LoadInt32(inits.Get(t)), 0)

				holder.Get(t).Finish()
				assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
				assert.Equal(t, atomic.LoadInt32(inits.Get(t)), 1)
			})

			s.And("the other test holds a different key", func(s *testcase.Spec) {
				holderKey.Let(s, func(t *testcase.T) string {
					return key.Get(t) + "-other"
				})

				s.Then("it doesn't wait", func(t *testcase.T) {
					assert.Within(t, syncTimeout, func(context.Context) { act(t) })
				})
			})

			s.And("the Var has a Before hook", func(s *testcase.Spec) {
				subject.Let(s, func(t *testcase.T) testcase.Var[int] {
					v := subject.Super(t)
					n := befores.Get(t)
					v.Before = func(*testcase.T, testcase.Var[int]) { atomic.AddInt32(n, 1) }
					return v
				})

				s.Then("the hook only runs once the test holds the key", func(t *testcase.T) {
					w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })
					assert.Equal(t, atomic.LoadInt32(befores.Get(t)), 0)

					holder.Get(t).Finish()
					assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
					assert.Equal(t, atomic.LoadInt32(befores.Get(t)), 1)
				})
			})

			s.And("the Var has dependencies", func(s *testcase.Spec) {
				// depInits counts the initialisations of the dependency
				depInits := let.Var(s, func(t *testcase.T) *int32 {
					return new(int32)
				})
				subject.Let(s, func(t *testcase.T) testcase.Var[int] {
					v := subject.Super(t)
					n := depInits.Get(t)
					v.Deps = testcase.Vars{testcase.Var[int]{
						ID: testcase.VarID(key.Get(t) + "-dep"),
						Init: func(t *testcase.T) int {
							atomic.AddInt32(n, 1)
							return t.Random.Int()
						},
					}}
					return v
				})

				s.Then("the dependencies are only initialised once the test holds the key", func(t *testcase.T) {
					w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })
					assert.Equal(t, atomic.LoadInt32(depInits.Get(t)), 0)

					holder.Get(t).Finish()
					assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
					assert.Equal(t, atomic.LoadInt32(depInits.Get(t)), 1)
				})
			})
		})

		s.When("the test already used the Var", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				_ = subject.Get(t).Get(tct.Get(t))
			})

			s.Then("it doesn't wait for itself", func(t *testcase.T) {
				assert.Within(t, syncTimeout, func(context.Context) { act(t) })
			})
		})

		s.When("the Var is bound to the Spec", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				subject.Get(t).Bind(spec.Get(t))
			})

			s.Then("it doesn't wait for itself", func(t *testcase.T) {
				assert.Within(t, syncTimeout, func(context.Context) { act(t) })
			})
		})

		s.When("the Var doesn't require synchronisation", func(s *testcase.Spec) {
			subject.Let(s, func(t *testcase.T) testcase.Var[int] {
				v := subject.Super(t)
				v.Sync = false
				return v
			})

			s.Then("other tests can use the Var's ID as key meanwhile", func(t *testcase.T) {
				act(t)
				other := newSyncTB(t, "TestOther")

				assert.Within(t, syncTimeout, func(context.Context) {
					testcase.Sync(other, key.Get(t))
				})
			})

			s.Then("other tests can use the Var at the same time", func(t *testcase.T) {
				act(t)

				assertConcurrentVarUse(t, tb.Get(t), func(other *testcase.FakeTB) {
					_ = subject.Get(t).Get(testcase.NewTWithSpec(other, testcase.NewSpec(other)))
				})
			})
		})
	})

	s.Describe("binding to a Spec", func(s *testcase.Spec) {
		var (
			bind = let.Var(s, func(t *testcase.T) func(testcase.Var[int], *testcase.Spec) {
				return func(v testcase.Var[int], sp *testcase.Spec) { v.Bind(sp) }
			})
		)
		// act binds the Var to the Spec, then starts a test of the Spec
		act := func(t *testcase.T) {
			bind.Get(t)(subject.Get(t), spec.Get(t))
			_ = tct.Get(t)
		}

		s.Then("the test holds the Var's ID as key from its start, even without using the Var", func(t *testcase.T) {
			act(t)
			other := newSyncTB(t, "TestOther")

			w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) {
				testcase.Sync(other, key.Get(t))
			})

			tb.Get(t).Finish()
			assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
			assert.Equal(t, atomic.LoadInt32(inits.Get(t)), 0, "the Var is still lazily initialised")
		})

		s.When("another test holds the key", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testcase.Sync(holder.Get(t), holderKey.Get(t))
			})

			s.Then("the test waits at its start, until the other test finishes", func(t *testcase.T) {
				w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })

				holder.Get(t).Finish()
				assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
			})

			s.And("the Var has a Before hook", func(s *testcase.Spec) {
				subject.Let(s, func(t *testcase.T) testcase.Var[int] {
					v := subject.Super(t)
					n := befores.Get(t)
					v.Before = func(*testcase.T, testcase.Var[int]) { atomic.AddInt32(n, 1) }
					return v
				})

				s.Then("the hook only runs once the test holds the key", func(t *testcase.T) {
					w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })
					assert.Equal(t, atomic.LoadInt32(befores.Get(t)), 0)

					holder.Get(t).Finish()
					assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
					assert.Equal(t, atomic.LoadInt32(befores.Get(t)), 1)
				})
			})
		})

		s.When("the Var is bound with Var.Let", func(s *testcase.Spec) {
			bind.Let(s, func(t *testcase.T) func(testcase.Var[int], *testcase.Spec) {
				return func(v testcase.Var[int], sp *testcase.Spec) { v.Let(sp, v.Init) }
			})

			s.Then("the test holds the Var's ID as key from its start", func(t *testcase.T) {
				act(t)
				other := newSyncTB(t, "TestOther")

				w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) {
					testcase.Sync(other, key.Get(t))
				})

				tb.Get(t).Finish()
				assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
			})
		})

		s.When("the Var is bound with Var.LetValue", func(s *testcase.Spec) {
			bind.Let(s, func(t *testcase.T) func(testcase.Var[int], *testcase.Spec) {
				value := t.Random.Int()
				return func(v testcase.Var[int], sp *testcase.Spec) { v.LetValue(sp, value) }
			})

			s.Then("the test holds the Var's ID as key from its start", func(t *testcase.T) {
				act(t)
				other := newSyncTB(t, "TestOther")

				w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) {
					testcase.Sync(other, key.Get(t))
				})

				tb.Get(t).Finish()
				assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
			})
		})

		s.When("the Var doesn't require synchronisation", func(s *testcase.Spec) {
			subject.Let(s, func(t *testcase.T) testcase.Var[int] {
				v := subject.Super(t)
				v.Sync = false
				return v
			})

			s.Then("other tests can use the Var's ID as key meanwhile", func(t *testcase.T) {
				act(t)
				other := newSyncTB(t, "TestOther")

				assert.Within(t, syncTimeout, func(context.Context) {
					testcase.Sync(other, key.Get(t))
				})
			})

			s.Then("other tests can use the Var at the same time", func(t *testcase.T) {
				act(t)

				assertConcurrentVarUse(t, tb.Get(t), func(other *testcase.FakeTB) {
					sp := testcase.NewSpec(other)
					bind.Get(t)(subject.Get(t), sp)
					_ = testcase.NewTWithSpec(other, sp)
				})
			})
		})
	})
}

// assertConcurrentVarUse asserts that other tests can use the Var while the test still uses it,
// and they can do it at the same time.
// The other tests keep using the Var, until all of them use it at the same time.
func assertConcurrentVarUse(t *testcase.T, test *testcase.FakeTB, use func(other *testcase.FakeTB)) {
	t.Helper()
	var (
		inUse  tcsync.Phaser
		others []*testcase.FakeTB
	)
	for i, n := 0, t.Random.IntBetween(2, 5); i < n; i++ {
		others = append(others, newSyncTB(t, fmt.Sprintf("TestOther%d", i)))
	}

	var blks []func()
	for _, other := range others {
		other := other
		blks = append(blks, func() {
			defer other.Finish()
			use(other)
			inUse.Wait()
		})
	}
	blks = append(blks, func() {
		// when the other tests can't use the Var at the same time,
		// finishing the tests lets them use it one by one, so the race can end.
		defer inUse.Finish()
		defer test.Finish()

		t.Eventually(func(t *testcase.T) {
			assert.Equal(t, inUse.Len(), len(others))
		})
	})
	testcase.Race(blks...)
}
