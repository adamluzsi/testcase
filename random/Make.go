package random

func (r *Random) Make(T any) any {
	return r.Factory.Make(r, T)
}
