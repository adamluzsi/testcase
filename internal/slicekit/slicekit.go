package slicekit

func ReverseLookup[T any](vs []T, index int) (T, bool) {
	return Lookup[T](vs, (-1*index)-1)
}

func Lookup[T any](vs []T, index int) (T, bool) {
	index, ok := normaliseIndex(len(vs), index)
	if !ok {
		var zero T
		return zero, false
	}
	return vs[index], true
}

// Merge will merge every []T slice into a single one.
func Merge[T any](slices ...[]T) []T {
	var out []T
	for _, slice := range slices {
		out = append(out, slice...)
	}
	return out
}

func normaliseIndex(length, index int) (int, bool) {
	if index < 0 {
		n := length + index
		return n, 0 <= n
	}
	return index, index < length
}
