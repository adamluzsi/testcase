package timecop

import "go.llib.dev/testcase/clock/internal"

type TravelOption interface {
	configure(option *internal.Option)
}

func toOption(tos []TravelOption) internal.Option {
	var o internal.Option
	for _, opt := range tos {
		opt.configure(&o)
	}
	return o
}

type fnTravelOption func(option *internal.Option)

func (fn fnTravelOption) configure(o *internal.Option) { fn(o) }
