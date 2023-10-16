package examples_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
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
		name = testcase.Let(s, func(t *testcase.T) string {
			return t.Random.String()
		})
		nameGet = func(t *testcase.T) string { return name.Get(t) }
		subject = func(t *testcase.T) string {
			return Say(nameGet(t))
		}
	)

	s.Then(`it will include the name`, func(t *testcase.T) {
		t.Must.Contain(subject(t), nameGet(t))
	})

	s.Then(`it should end the sentence with an exclamation mark`, func(t *testcase.T) {
		assert.Must(t).True(strings.HasSuffix(subject(t), `!`))
	})

	s.Then(`it should use one of the greeting`, func(t *testcase.T) {
		t.Must.Contain(Greetings, regexp.MustCompile(`([^\s]+)`).FindString(subject(t)))
	})
}
