package errorkitlite

import (
	"errors"
	"fmt"
	"strings"
)

const ErrNotImplemented Error = "ErrNotImplemented"

type Error string

func (err Error) Error() string { return string(err) }

func (err Error) F(format string, a ...any) error {
	return W{E: err, W: fmt.Errorf(format, a...)}
}

// Finish is a helper function that can be used from a deferred context.
//
// Usage:
//
//	defer errorkit.Finish(&returnError, rows.Close)
func Finish(returnErr *error, blk func() error) {
	*returnErr = Merge(*returnErr, blk())
}

// FinishOnError is a helper function that can be used from a deferred context.
// It runs the block conditionally, when the return error, which was assigned by the `return` keyword is not nil.
//
// Usage:
//
//	defer errorkit.FinishOnError(&returnError, func() { rollback(ctx) })
func FinishOnError(returnErr *error, blk func()) {
	if returnErr == nil || *returnErr == nil {
		return
	}
	blk()
}

// Recover will attempt a recover, and if recovery yields a value, it sets it as an error.
func Recover(returnErr *error) {
	var err = ToError(recover())
	if err != nil {
		*returnErr = Merge(*returnErr, err)
	}
}

func ToError(v any) error {
	if v == nil {
		return nil
	}
	if err, ok := v.(error); ok {
		return err
	}
	return fmt.Errorf("%v", v)
}

// As function serves as a shorthand to enable one-liner error handling with errors.As.
// It's meant to be used within an if statement, much like Lookup functions such as os.LookupEnv.
func As[T error](err error) (T, bool) {
	var v T
	ok := errors.As(err, &v)
	return v, ok
}

// Merge will combine all given non nil error values into a single error value.
// If no valid error is given, nil is returned.
// If only a single non nil error value is given, the error value is returned.
func Merge(errs ...error) error {
	var cleanErrs []error
	for _, err := range errs {
		if err == nil {
			continue
		}
		cleanErrs = append(cleanErrs, err)
	}
	errs = cleanErrs
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return MultiError(errs)
}

type MultiError []error

func (errs MultiError) Error() string {
	var msgs []string
	for _, err := range errs {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

func (errs MultiError) As(target any) bool {
	for _, err := range errs {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

func (errs MultiError) Is(target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// ErrFunc is a function that checks whether a stateful system currently has an error.
// For example context.Context#Err is an ErrFunc.
type ErrFunc = func() error

func NullErrFunc() error { return nil }

func MergeErrFunc(errFuncs ...ErrFunc) func() error {
	var fns []ErrFunc
	for _, fn := range errFuncs {
		if fn == nil {
			continue
		}
		fns = append(fns, ErrFunc(fn))
	}
	switch len(fns) {
	case 0:
		return NullErrFunc
	case 1:
		return ErrFunc(fns[0])
	}
	return func() (returnError error) {
		for i := len(fns) - 1; 0 <= i; i-- {
			defer Finish(&returnError, fns[i])
		}
		return nil
	}
}

type W struct {
	E Error
	W error
}

func (w W) Error() string {
	var msg string
	if w.W != nil {
		msg = w.W.Error()
	}
	return fmt.Sprintf("[%s] %s", w.E, msg)
}

func (w W) As(target any) bool {
	if errors.As(w.E, target) {
		return true
	}
	if errors.As(w.W, target) {
		return true
	}
	return false
}

func (w W) Is(target error) bool {
	if errors.Is(w.E, target) {
		return true
	}
	if errors.Is(w.W, target) {
		return true
	}
	return false
}
