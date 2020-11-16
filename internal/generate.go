//go:generate bash ./generate.sh
package internal

import "testing"

type TB interface {
	testing.TB
}
