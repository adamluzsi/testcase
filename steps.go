package testrunctx

import "testing"

type Steps []func(*testing.T)

func (s Steps) Setup(t *testing.T) {
	for _, step := range s {
		step(t)
	}
}
