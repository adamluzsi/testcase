package random

func Pick[T any](rnd *Random, vs ...T) T {
	if rnd == nil {
		rnd = defaultRandom
	}
	return rnd.SliceElement(vs).(T)
}
