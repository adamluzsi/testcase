package testcase

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestFullyQualifiedName(t *testing.T) {
	type Example struct{}
	var (
		rv           = reflect.TypeOf(Example{})
		expectedName = fmt.Sprintf(`%q.%s`, rv.PkgPath(), rv.Name())
	)
	require.Equal(t, expectedName, fullyQualifiedName(Example{}))
	require.Equal(t, expectedName, fullyQualifiedName(&Example{}))
}
