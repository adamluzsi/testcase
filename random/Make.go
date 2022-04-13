package random

func Make[T any](rnd *Random) T {
	var v T
	return rnd.Factory.Make(rnd, v).(T)
}

func (r *Random) Make(T any) any {
	return r.Factory.Make(r, T)
}
