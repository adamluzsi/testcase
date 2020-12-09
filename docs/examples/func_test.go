package examples_test

import (
	"fmt"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

// config
var Greetings = []string{`Hello`, `Hola`, `Howdy`}

// package mydomain

func Say(name string) string {
	return fmt.Sprintf(`%s %s!`, Greetings[rand.Intn(len(Greetings))], name)
}

// package mydomain test
func TestSay(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		name    = s.LetValue(`message as input`, fixtures.Random.String())
		nameGet = func(t *testcase.T) string { return name.Get(t).(string) }
		subject = func(t *testcase.T) string {
			return Say(nameGet(t))
		}
	)

	s.Then(`it will include the name`, func(t *testcase.T) {
		require.Contains(t, subject(t), nameGet(t))
	})

	s.Then(`it should end the sentence with an exclamation mark`, func(t *testcase.T) {
		require.True(t, strings.HasSuffix(subject(t), `!`))
	})

	s.Then(`it should use one of the greeting`, func(t *testcase.T) {
		require.Contains(t, Greetings, regexp.MustCompile(`([^\s]+)`).FindString(subject(t)))
	})
}
