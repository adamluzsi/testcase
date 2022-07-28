package random

func (r *Random) Make(T any) any {
	return r.Factory.Make(r, T)
}

func MakeSlice[T any](rnd *Random, length int) []T {
	var (
		typ T
		vs  []T
	)
	for i := 0; i < length; i++ {
		vs = append(vs, rnd.Make(typ).(T))
	}
	return vs
}

func MakeMap[K comparable, V any](rnd *Random, length int) map[K]V {
	var (
		kT K
		vT V
		vs = make(map[K]V)
	)
	for i := 0; i < length; i++ {
		k := rnd.Make(kT).(K)
		v := rnd.Make(vT).(V)
		if _, ok := vs[k]; ok {
			i++
			continue
		}
		vs[k] = v
	}
	return vs
}
