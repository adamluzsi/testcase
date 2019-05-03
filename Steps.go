package testcase

import "testing"

type Steps []func(*testing.T)

func (s Steps) Add(step func(*testing.T)) Steps {
	return append(append(Steps{}, s...), step)
}

func (s Steps) Setup(t *testing.T) {
	for _, steps := range s {
		steps(t)
	}
}
