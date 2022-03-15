package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
)

type MessageWrapper struct {
	Message string
}

func (mt MessageWrapper) LookupMessage() (string, bool) {
	if mt.Message == `` {
		return ``, false
	}

	return mt.Message, true
}

func TestMessageWrapper(t *testing.T) {
	s := testcase.NewSpec(t)
	s.NoSideEffect()

	var (
		message        = testcase.Var[string]{ID: `message`}
		messageWrapper = testcase.Let(s, func(t *testcase.T) MessageWrapper {
			return MessageWrapper{Message: message.Get(t)}
		})
	)

	s.Describe(`#LookupMessage`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (string, bool) {
			return messageWrapper.Get(t).LookupMessage()
		}

		s.When(`message is empty`, func(s *testcase.Spec) {
			message.LetValue(s, ``)

			s.Then(`it will return with "ok" as false`, func(t *testcase.T) {
				_, ok := subject(t)
				t.Must.True(!ok)
			})
		})

		s.When(`message has content`, func(s *testcase.Spec) {
			message.LetValue(s, fixtures.Random.String())

			s.Then(`it will return with "ok" as true`, func(t *testcase.T) {
				_, ok := subject(t)
				t.Must.True(ok)
			})

			s.Then(`message received back`, func(t *testcase.T) {
				msg, _ := subject(t)
				t.Must.Equal(message.Get(t), msg)
			})
		})
	})
}

func ExampleSpec() {
	var t *testing.T
	TestMessageWrapper(t)
}
