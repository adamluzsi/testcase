package random

func (r *Random) Make(T any) any {
	return r.Factory.Make(r, T)
}

func Slice[T any](length int, mk func() T, opts ...sliceOption) []T {
	var c sliceConfig
	c.use(opts)
	var vs []T
	for i := 0; i < length; i++ {
		var v T
		if c.Unique {
			v = Unique(mk, vs...)
		} else {
			v = mk()
		}
		vs = append(vs, v)
	}
	return vs
}

func Map[K comparable, V any](length int, mk func() (K, V), opts ...mapOption) map[K]V {
	var c mapConfig
	c.use(opts)
	var (
		vs               = make(map[K]V)
		collisionRetries = 42
	)
	for i := 0; i < length; i++ {
		var (
			k K
			v V
		)
		if c.Unique {
			var vals []V
			for _, val := range vs {
				vals = append(vals, val)
			}
			Unique(func() V {
				k, v = mk()
				return v
			}, vals...)
		} else {
			k, v = mk()
		}
		if _, ok := vs[k]; ok {
			if 0 < collisionRetries {
				collisionRetries--
				i--
			}
			continue
		}
		vs[k] = v
	}
	return vs
}

func KV[K comparable, V any](mkK func() K, mkV func() V) func() (K, V) {
	return func() (K, V) {
		return mkK(), mkV()
	}
}

type sliceConfig struct {
	Unique bool
}

func (c *sliceConfig) use(opts []sliceOption) {
	for _, opt := range opts {
		opt.sliceOption(c)
	}
}

type sliceOption interface {
	sliceOption(*sliceConfig)
}

type mapConfig struct {
	Unique bool
}

func (c *mapConfig) use(opts []mapOption) {
	for _, opt := range opts {
		opt.mapOption(c)
	}
}

type mapOption interface {
	mapOption(*mapConfig)
}
