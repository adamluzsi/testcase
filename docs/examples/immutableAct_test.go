package examples_test

import (
	"strings"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/docs/examples"
)

func TestImmutableAct(t *testing.T) {
	s := testcase.NewSpec(t)

	myStruct := testcase.Let(s, func(t *testcase.T) *examples.MyStruct {
		return &examples.MyStruct{}
	})

	s.Describe(`#Shrug`, func(s *testcase.Spec) {
		const shrugEmoji = `¯\_(ツ)_/¯`
		var (
			message = testcase.Let(s, func(t *testcase.T) string { return t.Random.String() })
			subject = func(t *testcase.T) string {
				return myStruct.Get(t).Shrug(message.Get(t))
			}
		)

		s.When(`message doesn't have shrug in the ending`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				t.Must.Contains(subject(t), shrugEmoji)
			})

			s.Then(`it will append shrug emoji to this`, func(t *testcase.T) {
				t.Must.True(strings.HasSuffix(subject(t), shrugEmoji))
			})
		})

		s.When(`shrug part of the input message`, func(s *testcase.Spec) {
			message.Let(s, func(t *testcase.T) string {
				return t.Random.String() + shrugEmoji
			})

			s.Then(`it will not append any more shrug emoji to the end of the message`, func(t *testcase.T) {
				t.Must.Equal(message.Get(t), subject(t))
			})
		})
	})
}
