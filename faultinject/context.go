package faultinject

import (
	"context"
	"path"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/adamluzsi/testcase/internal/caller"
)

// Inject will inject the received fault into a context.
// Check calls that match the Fault's requirements will yield error.
func Inject(ctx context.Context, fault Fault) context.Context {
	if v, ok := lookup(ctx); ok {
		*v = append(*v, fault)
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, &[]Fault{fault})
}

// Check will check whether the given context contains Fault which should be returned.
// If Check returns an error because an injected Fault, the Fault is consumed and won't happen again.
// Using Check allows you to inject faults without using mocks and indirections.
// By default, Check will return quickly in case there is no fault injection present.
func Check(ctx context.Context, tags ...string) error {
	fs, ok := lookup(ctx)
	if !ok { // quick path
		return nil
	}
	format := func(fname string) string {
		base := path.Base(filepath.Base(fname))
		return matchFuncSuffix.ReplaceAllString(base, "")
	}
	frame, hasFrame := caller.MatchFrame(func(frame runtime.Frame) bool {
		return matchFuncName.MatchString(format(frame.Function))
	})
	var funcName string
	if hasFrame {
		funcName = matchFuncName.FindString(format(frame.Function))
	}
	fault, ok := nextFault(fs, func(fault Fault) bool {
		if hasFrame && fault.OnFunc != "" && fault.OnFunc == funcName {
			return true
		}
		for _, tag := range tags {
			if fault.OnTag == tag {
				return true
			}
		}
		return false
	})
	if !ok {
		return nil
	}
	return fault.Error
}

type ctxKey struct{}

func lookup(ctx context.Context) (*[]Fault, bool) {
	if ctx == nil {
		return nil, false
	}
	v, ok := ctx.Value(ctxKey{}).(*[]Fault)
	return v, ok
}

var (
	matchFuncName   = regexp.MustCompile(`^[^\.]+\.[^\.]+(?:\.[^\.]+)?`)
	matchFuncSuffix = regexp.MustCompile(`\.func\d+.*$`)
)
