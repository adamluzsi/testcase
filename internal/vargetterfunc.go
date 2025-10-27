package internal

type VarGetterFunc[T any, V any] func(t *T) V

func (fn VarGetterFunc[T, V]) Get(t *T) V { return fn(t) }
