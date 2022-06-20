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
	fs, ok := lookup(ctx)
	if !ok { // quick path
		return nil
	}
	err, ok := i.next(fs, func(nt Tag) (error, bool) {
		for tag, error := range i.cases {
			if nt == tag {
				return error, true
			}
		}
		return nil, false
	})
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
