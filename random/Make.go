package random

func (r *Random) Make(T any) any {
	return r.Factory.Make(r, T)
}

func Slice[T any](length int, mk func() T) []T {
	var vs []T
	for i := 0; i < length; i++ {
		vs = append(vs, mk())
	}
	return vs
}

func Map[K comparable, V any](length int, mk func() (K, V)) map[K]V {
	var (
		vs               = make(map[K]V)
		collisionRetries = 42
	)
	for i := 0; i < length; i++ {
		k, v := mk()
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
