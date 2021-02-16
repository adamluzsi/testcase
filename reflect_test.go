package testcase

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFullyQualifiedName(t *testing.T) {
	type Example struct{}
	var (
		rv           = reflect.TypeOf(Example{})
		expectedName = fmt.Sprintf(`%q.%s`, rv.PkgPath(), rv.Name())
	)
	require.Equal(t, expectedName, fullyQualifiedName(Example{}))
	require.Equal(t, expectedName, fullyQualifiedName(&Example{}))
	require.Equal(t, `string`, fullyQualifiedName("42"))
}
