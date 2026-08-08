package errorkitlite_test

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/errorkitlite"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
)

var rnd = random.New(random.CryptoSeed{})

type (
	ErrType1 struct{}
	ErrType2 struct{ V int }
)

func (err ErrType1) Error() string { return "ErrType1" }
func (err ErrType2) Error() string { return "ErrType2" }

type MyError struct {
	Msg string
}

func (err MyError) Error() string {
	return err.Msg
}

func TestMerge(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		errs = testcase.Let[[]error](s, nil)
	)
	act := func(t *testcase.T) error {
		return errorkitlite.Merge(errs.Get(t)...)
	}

	s.When("no error is supplied", func(s *testcase.Spec) {
		errs.Let(s, func(t *testcase.T) []error {
			return []error{}
		})

		s.Then("it will return with nil", func(t *testcase.T) {
			assert.Must(t).NoError(act(t))
		})

		s.Then("errors.Is yield false", func(t *testcase.T) {
			err := act(t)
			assert.Must(t).False(errors.Is(err, ErrType1{}))
			assert.Must(t).False(errors.Is(err, ErrType2{}))
		})

		s.Then("errors.As yield false", func(t *testcase.T) {
			err := act(t)
			assert.Must(t).False(errors.As(err, &ErrType1{}))
			assert.Must(t).False(errors.As(err, &ErrType2{}))
		})
	})

	s.When("an error value is supplied", func(s *testcase.Spec) {
		expectedErr := let.Error(s)

		errs.Let(s, func(t *testcase.T) []error {
			return []error{expectedErr.Get(t)}
		})

		s.Then("the exact value is returned", func(t *testcase.T) {
			assert.Must(t).Equal(expectedErr.Get(t), act(t))
		})

		s.Then("errors.Is yield false", func(t *testcase.T) {
			err := act(t)
			assert.Must(t).False(errors.Is(err, ErrType1{}))
			assert.Must(t).False(errors.Is(err, ErrType2{}))
		})

		s.Then("errors.As yield false", func(t *testcase.T) {
			err := act(t)
			assert.Must(t).False(errors.As(err, &ErrType1{}))
			assert.Must(t).False(errors.As(err, &ErrType2{}))
		})

		s.And("the error value is a typed error value", func(s *testcase.Spec) {
			expectedErr.LetValue(s, ErrType1{})

			s.Then("the exact value is returned", func(t *testcase.T) {
				assert.Must(t).Equal(expectedErr.Get(t), act(t))
			})

			s.Then("errors.Is will find wrapped error", func(t *testcase.T) {
				err := act(t)
				assert.True(t, errors.Is(err, ErrType1{}))
				assert.Must(t).False(errors.Is(err, ErrType2{}))
			})

			s.Then("errors.As will find the wrapped error", func(t *testcase.T) {
				err := act(t)
				assert.True(t, errors.As(err, &ErrType1{}))
				assert.Must(t).False(errors.As(err, &ErrType2{}))
			})
		})

		s.And("but the error value is nil", func(s *testcase.Spec) {
			expectedErr.LetValue(s, nil)

			s.Then("it will return with nil", func(t *testcase.T) {
				assert.Must(t).NoError(act(t))
			})

			s.Then("errors.Is yield false", func(t *testcase.T) {
				err := act(t)
				assert.Must(t).False(errors.Is(err, ErrType1{}))
				assert.Must(t).False(errors.Is(err, ErrType2{}))
			})

			s.Then("errors.As yield false", func(t *testcase.T) {
				err := act(t)
				assert.Must(t).False(errors.As(err, &ErrType1{}))
				assert.Must(t).False(errors.As(err, &ErrType2{}))
			})
		})
	})

	s.When("multiple error values are supplied", func(s *testcase.Spec) {
		expectedErr1 := let.Error(s)
		expectedErr2 := let.Error(s)
		expectedErr3 := let.Error(s)

		errs.Let(s, func(t *testcase.T) []error {
			return []error{
				expectedErr1.Get(t),
				expectedErr2.Get(t),
				expectedErr3.Get(t),
			}
		})

		s.Then("retruned value includes all three error value", func(t *testcase.T) {
			err := act(t)
			assert.Must(t).ErrorIs(expectedErr1.Get(t), err)
			assert.Must(t).ErrorIs(expectedErr2.Get(t), err)
			assert.Must(t).ErrorIs(expectedErr2.Get(t), err)
		})

		s.Then("errors.Is yield false", func(t *testcase.T) {
			err := act(t)
			assert.Must(t).False(errors.Is(err, ErrType1{}))
			assert.Must(t).False(errors.Is(err, ErrType2{}))
		})

		s.Then("errors.As yield false", func(t *testcase.T) {
			err := act(t)
			assert.Must(t).False(errors.As(err, &ErrType1{}))
			assert.Must(t).False(errors.As(err, &ErrType2{}))
		})

		s.And("the errors has a typed error value", func(s *testcase.Spec) {
			expectedErr2.LetValue(s, ErrType1{})

			s.Then("the named error value is returned", func(t *testcase.T) {
				assert.Must(t).ErrorIs(expectedErr2.Get(t), act(t))
			})

			s.Then("errors.Is can find the wrapped error", func(t *testcase.T) {
				err := act(t)
				assert.True(t, errors.Is(err, ErrType1{}))
				assert.Must(t).False(errors.Is(err, ErrType2{}))
			})

			s.Then("errors.As can find the wrapped error", func(t *testcase.T) {
				err := act(t)
				assert.True(t, errors.As(err, &ErrType1{}))
				assert.Must(t).False(errors.As(err, &ErrType2{}))
			})
		})

		s.And("the errors has multiple typed error value", func(s *testcase.Spec) {
			expectedErr2.LetValue(s, ErrType1{})
			expectedErr3.Let(s, func(t *testcase.T) error {
				return ErrType2{V: t.Random.Int()}
			})

			s.Then("returned error contains all typed error", func(t *testcase.T) {
				assert.Must(t).ErrorIs(expectedErr2.Get(t), act(t))
				assert.Must(t).ErrorIs(expectedErr3.Get(t), act(t))
			})

			s.Then("errors.Is can find the wrapped error", func(t *testcase.T) {
				err := act(t)
				assert.True(t, errors.Is(err, expectedErr2.Get(t)))
				assert.True(t, errors.Is(err, expectedErr3.Get(t)))
				assert.Must(t).False(errors.Is(err, ErrType2{}))
			})

			s.Then("errors.As can find the wrapped error", func(t *testcase.T) {
				err := act(t)
				assert.True(t, errors.As(err, &ErrType1{}))

				var gotErrWithAs ErrType2
				assert.True(t, errors.As(err, &gotErrWithAs))
				assert.Must(t).NotNil(gotErrWithAs)
				assert.Must(t).Equal(expectedErr3.Get(t), gotErrWithAs)
			})
		})

		s.And("but the error values are nil", func(s *testcase.Spec) {
			expectedErr1.LetValue(s, nil)
			expectedErr2.LetValue(s, nil)
			expectedErr3.LetValue(s, nil)

			s.Then("it will return with nil", func(t *testcase.T) {
				assert.Must(t).NoError(act(t))
			})

			s.Then("errors.Is yield false", func(t *testcase.T) {
				err := act(t)
				assert.Must(t).False(errors.Is(err, ErrType1{}))
				assert.Must(t).False(errors.Is(err, ErrType2{}))
			})

			s.Then("errors.As yield false", func(t *testcase.T) {
				err := act(t)
				assert.Must(t).False(errors.As(err, &ErrType1{}))
				assert.Must(t).False(errors.As(err, &ErrType2{}))
			})
		})
	})
}

func TestMergeErrFunc(t *testing.T) {
	s := testcase.NewSpec(t)

	var errFuncs = let.Var[[]errorkitlite.ErrFunc](s, nil)
	act := func(t *testcase.T) errorkitlite.ErrFunc {
		return errorkitlite.MergeErrFunc(errFuncs.Get(t)...)
	}

	s.When("no function is passed to it", func(s *testcase.Spec) {
		errFuncs.LetValue(s, nil)

		s.Then("a non-nil null object is returned", func(t *testcase.T) {
			got := act(t)
			assert.NotNil(t, got)
			assert.NotPanic(t, func() {
				assert.NoError(t, got())
			})
		})
	})

	s.When("single function is passed", func(s *testcase.Spec) {
		err := let.Error(s)

		fn := let.Var(s, func(t *testcase.T) errorkitlite.ErrFunc {
			return func() error { return err.Get(t) }
		})

		errFuncs.Let(s, func(t *testcase.T) []errorkitlite.ErrFunc {
			return []errorkitlite.ErrFunc{fn.Get(t)}
		})

		s.Then("the single function is returned as is", func(t *testcase.T) {
			got := act(t)
			assert.NotNil(t, got)
			v := fn.Get(t)
			assert.Equal(t, &got, &v)
			assert.Equal(t, err.Get(t), got())
		})
	})

	s.When("multiple ErrFunc passed to it", func(s *testcase.Spec) {
		var (
			fn1 = let.Var(s, func(t *testcase.T) func() error {
				var err = t.Random.Error()
				return func() error { return err }
			})
			fn2 = let.Var(s, func(t *testcase.T) func() error {
				var err = t.Random.Error()
				return func() error { return err }
			})
			fn3 = let.Var(s, func(t *testcase.T) func() error {
				var err = t.Random.Error()
				return func() error { return err }
			})
		)

		errFuncs.Let(s, func(t *testcase.T) []errorkitlite.ErrFunc {
			return []errorkitlite.ErrFunc{
				fn1.Get(t),
				fn2.Get(t),
				fn3.Get(t),
			}
		})

		s.Then("the error functions are merged along with the error value they would return", func(t *testcase.T) {
			got := act(t)
			assert.NotNil(t, got)

			gotErr := got()

			assert.ErrorIs(t, gotErr, fn1.Get(t)())
			assert.ErrorIs(t, gotErr, fn2.Get(t)())
			assert.ErrorIs(t, gotErr, fn3.Get(t)())
		})

		s.Then("the returned ErrFunc is idempotent with its merging process", func(t *testcase.T) {
			got := act(t)
			assert.NotNil(t, got)

			t.Random.Repeat(3, 6, func() {
				gotErr := got()
				assert.ErrorIs(t, gotErr, fn1.Get(t)())
				assert.ErrorIs(t, gotErr, fn2.Get(t)())
				assert.ErrorIs(t, gotErr, fn3.Get(t)())
			})
		})

		s.And("if all the functions are nil values", func(s *testcase.Spec) {
			fn1.LetValue(s, nil)
			fn2.LetValue(s, nil)
			fn3.LetValue(s, nil)

			s.Then("a non-nil null object is returned", func(t *testcase.T) {
				got := act(t)
				assert.NotNil(t, got)
				assert.NotPanic(t, func() {
					assert.NoError(t, got())
				})
			})
		})

		s.And("if one of the functions are a nil", func(s *testcase.Spec) {
			fn2.LetValue(s, nil)

			s.Then("we get back a merged ErrFunc from the rest of the non nil ErrFuncs", func(t *testcase.T) {
				got := act(t)
				assert.NotNil(t, got)

				gotErr := got()
				assert.ErrorIs(t, gotErr, fn1.Get(t)())
				assert.ErrorIs(t, gotErr, fn3.Get(t)())
			})
		})
	})

	s.Test("support func() error type", func(t *testcase.T) {
		expErr := t.Random.Error()

		var errFuncs []func() error
		errFuncs = append(errFuncs, func() error { return expErr })

		got := errorkitlite.MergeErrFunc(errFuncs...)
		assert.NotNil(t, got)
		assert.ErrorIs(t, got(), expErr)
	})
}

func TestFinish(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})

	t.Run("errors are merged from all source", func(t *testing.T) {
		err1 := rnd.Error()
		err2 := rnd.Error()

		got := func() (rErr error) {
			defer errorkitlite.Finish(&rErr, func() error {
				return err1
			})

			return err2
		}()

		assert.ErrorIs(t, err1, got)
		assert.ErrorIs(t, err2, got)
	})

	t.Run("Finish error is returned", func(t *testing.T) {
		exp := rnd.Error()
		got := func() (rErr error) {
			defer errorkitlite.Finish(&rErr, func() error {
				return exp
			})

			return nil
		}()

		assert.ErrorIs(t, exp, got)
	})

	t.Run("func return value returned", func(t *testing.T) {
		exp := rnd.Error()
		got := func() (rErr error) {
			defer errorkitlite.Finish(&rErr, func() error {
				return nil
			})

			return exp
		}()

		assert.ErrorIs(t, exp, got)
	})

	t.Run("nothing fails, no error returned", func(t *testing.T) {
		got := func() (rErr error) {
			defer errorkitlite.Finish(&rErr, errorkitlite.NullErrFunc)

			return nil
		}()

		assert.NoError(t, got)
	})
}

func TestRecover(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		action    = testcase.Var[func() error]{ID: `action`}
		actionLet = func(s *testcase.Spec, fn func() error) { action.Let(s, func(t *testcase.T) func() error { return fn }) }
	)
	subject := func(t *testcase.T) (err error) {
		defer errorkitlite.Recover(&err)
		return action.Get(t)()
	}

	s.When(`action ends without error`, func(s *testcase.Spec) {
		actionLet(s, errorkitlite.NullErrFunc)

		s.Then(`it will do nothing`, func(t *testcase.T) {
			assert.Must(t).NoError(subject(t))
		})
	})

	s.When(`action returns an error`, func(s *testcase.Spec) {
		expectedErr := errors.New(`boom`)
		actionLet(s, func() error { return expectedErr })

		s.Then(`it will pass the received error through`, func(t *testcase.T) {
			assert.Equal(t, expectedErr, subject(t))
		})
	})

	s.When(`action panics with an error`, func(s *testcase.Spec) {
		expectedErr := errors.New(`boom`)
		actionLet(s, func() error { panic(expectedErr) })

		s.Then(`it will capture the error from panic and returns with it`, func(t *testcase.T) {
			assert.Equal(t, expectedErr, subject(t))
		})
	})

	s.When(`action panics with an error`, func(s *testcase.Spec) {
		expectedErr := errors.New(`boom`)
		actionLet(s, func() error { panic(expectedErr) })

		s.Then(`it will capture the error from panic and returns with it`, func(t *testcase.T) {
			assert.Equal(t, expectedErr, subject(t))
		})
	})

	s.When(`action panics with an non error type`, func(s *testcase.Spec) {
		const msg = `boom`
		actionLet(s, func() error { panic(msg) })

		s.Then(`it will capture the panic value and create an error from it, where message is the panic object is formatted with fmt`, func(t *testcase.T) {
			assert.Equal(t, errors.New("boom"), subject(t))
		})
	})

	s.When(`action stops the go routine`, func(s *testcase.Spec) {
		actionLet(s, func() error {
			runtime.Goexit()
			return nil
		})

		s.Then(`it will let go exit continues`, func(t *testcase.T) {
			var (
				wg       = &sync.WaitGroup{}
				finished bool
			)
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = subject(t)
				finished = true
			}()
			wg.Wait()
			assert.Must(t).False(finished)
		})
	})
}

func BenchmarkFinish(b *testing.B) {
	var err error
	for i := 0; i < b.N; i++ {
		errorkitlite.Finish(&err, func() error {
			return nil
		})
	}
}

func ExampleAs() {
	var err error // some error to be checked

	if err, ok := errorkitlite.As[MyError](err); ok {
		fmt.Println(err.Msg)
	}
}

func TestAs(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		expErr := MyError{Msg: rnd.Error().Error()}
		var err error = fmt.Errorf("wrapped: %w", expErr)
		gotErr, ok := errorkitlite.As[MyError](err)
		assert.True(t, ok)
		assert.Equal(t, gotErr, expErr)
	})
	t.Run("rainy", func(t *testing.T) {
		var err error = fmt.Errorf("wrapped: %w", rnd.Error())
		gotErr, ok := errorkitlite.As[MyError](err)
		assert.False(t, ok)
		assert.Empty(t, gotErr)
	})
	t.Run("nil", func(t *testing.T) {
		gotErr, ok := errorkitlite.As[MyError](nil)
		assert.False(t, ok)
		assert.Empty(t, gotErr)
	})
}

func TestError(t *testing.T) {
	var err errorkitlite.Error = "foo/bar/baz"
	exp := errorkitlite.Error("foo/bar/baz")
	assert.ErrorIs(t, err, exp)
	assert.True(t, errors.Is(err, exp))
	assert.Equal(t, err.Error(), "foo/bar/baz")
}

func TestError_F_smoke(t *testing.T) {
	const ErrExample errorkitlite.Error = "ErrExample"
	t.Run("sprintf", func(t *testing.T) {
		got := ErrExample.F("foo - bar - %s", "baz")
		assert.ErrorIs(t, got, ErrExample)
		assert.Contains(t, got.Error(), "foo - bar - baz")
	})
	t.Run("errorf", func(t *testing.T) {
		exp := rnd.Error()
		got := ErrExample.F("%w", exp)
		assert.ErrorIs(t, got, ErrExample)
		assert.ErrorIs(t, got, exp)
		assert.Contains(t, got.Error(), ErrExample.Error())
	})
}

type StubErr struct {
	V int
}

func (err StubErr) Error() string {
	return strconv.Itoa(err.V)
}

func TestW_smoke(t *testing.T) {
	cstErr := errorkitlite.Error(rnd.Error().Error())
	expErr := StubErr{V: rnd.Int()}

	w := errorkitlite.W{E: cstErr, W: expErr}

	assert.Error(t, w)
	assert.Equal(t, w.Error(), fmt.Sprintf("[%s] %s", cstErr.Error(), expErr.Error()))

	assert.True(t, errors.Is(w, expErr))
	assert.True(t, errors.Is(w, cstErr))

	var gotCST errorkitlite.Error
	assert.True(t, errors.As(w, &gotCST))
	assert.Equal(t, gotCST, cstErr)

	var gotEXP StubErr
	assert.True(t, errors.As(w, &gotEXP))
	assert.Equal(t, expErr, gotEXP)
}

func ExampleFinishOnError() {
	// example function
	var _ = func() (rerr error) {
		defer errorkitlite.FinishOnError(&rerr, func() { /* do something when the */ })

		return errorkitlite.Error("boom")
	}
}

func TestFinishOnError(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		returnErr = testcase.Let[error](s, nil)
		blockRan  = testcase.LetValue(s, false)
	)
	act := func(t *testcase.T) {
		var v = returnErr.Get(t)
		defer errorkitlite.FinishOnError(&v, func() { blockRan.Set(t, true) })
	}

	s.When("return error is nil", func(s *testcase.Spec) {
		returnErr.LetValue(s, nil)

		s.Then("it will not execute the block", func(t *testcase.T) {
			act(t)

			assert.False(t, blockRan.Get(t))
		})
	})

	s.When("return was a valid error value", func(s *testcase.Spec) {
		returnErr.Let(s, func(t *testcase.T) error {
			return t.Random.Error()
		})

		s.Then("it will execute the block", func(t *testcase.T) {
			act(t)

			assert.True(t, blockRan.Get(t))
		})
	})
}

func TestToError(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		v = let.Var[any](s, nil)
	)
	act := let.Act(func(t *testcase.T) error {
		return errorkitlite.ToError(v.Get(t))
	})

	s.When("the value is nil", func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) any {
			return nil
		})

		s.Then("no error is returned", func(t *testcase.T) {
			assert.NoError(t, act(t))
		})
	})

	s.When("the value is an error", func(s *testcase.Spec) {
		expErr := let.Var(s, func(t *testcase.T) error {
			return t.Random.Error()
		})
		v.Let(s, func(t *testcase.T) any {
			return expErr.Get(t)
		})

		s.Then("the same error value is returned as-is", func(t *testcase.T) {
			got := act(t)

			assert.Equal(t, expErr.Get(t), got)
			assert.ErrorIs(t, got, expErr.Get(t))
		})

		s.And("the error is of a concrete custom type", func(s *testcase.Spec) {
			myError := let.Var(s, func(t *testcase.T) MyError {
				return MyError{Msg: t.Random.String()}
			})
			expErr.Let(s, func(t *testcase.T) error {
				return myError.Get(t)
			})

			s.Then("the concrete error value is returned", func(t *testcase.T) {
				got := act(t)

				assert.Equal(t, expErr.Get(t), got)
				assert.Equal[error](t, myError.Get(t), got)
				assert.ErrorIs(t, got, expErr.Get(t))
			})

			s.Then("the concrete error value is preserved, so errors.As keeps working", func(t *testcase.T) {
				got := act(t)

				var gotErr MyError
				assert.True(t, errors.As(got, &gotErr))
				assert.Equal(t, myError.Get(t), gotErr)
			})
		})
	})

	s.When("the value is not an error", func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) any {
			return t.Random.String()
		})

		s.Then("an error is built from the value's default formatting", func(t *testcase.T) {
			got := act(t)

			assert.Error(t, got)
			assert.Equal(t, fmt.Sprintf("%v", v.Get(t)), got.Error())
		})

		s.And("the value is a non-string type", func(s *testcase.Spec) {
			v.Let(s, func(t *testcase.T) any {
				return t.Random.Int()
			})

			s.Then("the error message is the fmt formatted value", func(t *testcase.T) {
				got := act(t)

				assert.Error(t, got)
				assert.Equal(t, fmt.Sprintf("%v", v.Get(t)), got.Error())
			})
		})
	})
}
