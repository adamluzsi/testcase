package mypkg

import "strings"

type MyType struct {
	ToUpper bool
}

func (mt MyType) MyFunc(v string) string {
	if mt.ToUpper {
		return strings.ToUpper(v)
	}
	return v
}
