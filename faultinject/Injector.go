package faultinject

import (
	"context"
)

type Injector struct {
	cases map[string]error
}

func (i Injector) OnTag(tag string, err error) Injector {
	if i.cases == nil {
		i.cases = make(map[string]error)
	}
	if _, ok := i.cases[tag]; !ok {
		i.cases[tag] = err
	}
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
	_, err, ok := i.next(fs, func(fault fault) (error, bool) {
		for tag, error := range i.cases {
			if fault.Tag == tag {
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

func (i Injector) next(fs *[]fault, filter func(fault) (error, bool)) (fault, error, bool) {
	var (
		nfs   = make([]fault, 0, len(*fs))
		fault fault
		rerr  error
		ok    bool
	)
	for _, f := range *fs {
		if !ok {
			if err, has := filter(f); has {
				fault = f
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
	return fault, rerr, ok
}
