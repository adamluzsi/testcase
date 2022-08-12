package timecop

type TravelOption interface {
	configure(*option)
}

func toOption(tos []TravelOption) option {
	var o option
	for _, opt := range tos {
		opt.configure(&o)
	}
	return o
}

type fnTravelOption func(*option)

func (fn fnTravelOption) configure(o *option) { fn(o) }

type option struct {
	Freeze bool
}
