package testrunctx

import "testing"

func Setup(t *testing.T, steps ...func(*testing.T)) {
	for _, step := range steps {
		step(t)
	}
}
