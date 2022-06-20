package faultinject

import (
	"context"
)

type Injector struct {
	cases InjectorCases
}

type InjectorCases map[Tag]error

func (i Injector) OnTag(tag Tag, err error) Injector {
	return i.OnTags(InjectorCases{tag: err})
}

func (i Injector) OnTags(newCases InjectorCases) Injector {
	cases := make(InjectorCases)
	for ctag, cErr := range i.cases {
		cases[ctag] = cErr
	}
	for ctag, cErr := range newCases {
		cases[ctag] = cErr
	}
	i.cases = cases
	return i
}

// Check will check whether the given context contains fault which should be returned.
// If Check returns an error because an injected fault, the fault is consumed and won't happen again.
// Using Check allows you to inject faults without using mocks and indirections.
// By default, Check will return quickly in case there is no fault injection present.
func (i Injector) Check(ctx context.Context) error {
	return i.check(ctx, func(tag Tag) (error, bool) {
		for targetTag, err := range i.cases {
			if tag == targetTag {
				return err, true
			}
		}
		return nil, false
	})
}

// CheckFor will check if the target tag has a fault injected into the context.
// It is ideal if you want to use a single Injector, but check for individual faults.
func (i Injector) CheckFor(ctx context.Context, target Tag) error {
	return i.check(ctx, func(tag Tag) (error, bool) {
		if tag != target {
			return nil, false
		}
		err, ok := i.cases[target]
		return err, ok
	})
}

func (i Injector) check(ctx context.Context, filter func(Tag) (error, bool)) error {
	if !Enabled {
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
