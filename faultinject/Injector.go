package faultinject

import (
	"context"
)

type Injector struct {
	cases InjectorCases
}

type InjectorCases map[Tag]error

func (i Injector) OnTag(tag Tag, err error) Injector {
	cases := make(InjectorCases)
	for ctag, cErr := range i.cases {
		cases[ctag] = cErr
	}
	cases[tag] = err
	i.cases = cases
	return i
}

// Check will Check whether the given context contains fault which should be returned.
// If Check returns an error because an injected fault, the fault is consumed and won't happen again.
// Using Check allows you to inject faults without using mocks and indirections.
// By default, Check will return quickly in case there is no fault injection present.
func (i Injector) Check(ctx context.Context) error {
	return i.check(ctx, func(fault Tag) (error, bool) {
		if err, ok := i.asFault(fault); ok {
			return err, ok
		}
		for targetFault, err := range i.cases {
			if fault == targetFault {
				return err, true
			}
		}
		return nil, false
	})
}

// CheckFor will Check if the target tag has a fault injected into the context.
// It is ideal if you want to use a single Injector, but Check for individual faults.
func (i Injector) CheckFor(ctx context.Context, target Tag) error {
	return i.check(ctx, func(tag Tag) (error, bool) {
		if err, ok := i.asFault(tag); ok {
			return err, ok
		}
		if tag != target {
			return nil, false
		}
		err, ok := i.cases[target]
		return err, ok
	})
}

func (i Injector) check(ctx context.Context, filter func(Tag) (error, bool)) error {
	if !Enabled() {
		return nil
	}
	fs, ok := lookup(ctx)
	if !ok { // quick path
		return nil
	}
	err, ok := i.next(fs, filter)
	if !ok {
		return nil
	}
	return err
}

func (i Injector) next(fs *[]Tag, filter func(Tag) (error, bool)) (error, bool) {
	var (
		nfs  = make([]Tag, 0, len(*fs))
		rerr error
		ok   bool
	)
	for _, f := range *fs {
		if !ok {
			if err, has := filter(f); has {
				ok = true
				rerr = err
				continue
			}
		}
		nfs = append(nfs, f)
	}
	if ok {
		*fs = nfs
	}
	return rerr, ok
}

func (i Injector) asFault(tag Tag) (error, bool) {
	fault, ok := tag.(Fault)
	if !ok {
		return nil, false
	}
	return fault.check()
}
