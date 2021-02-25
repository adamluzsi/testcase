package examples_test

import (
	"strings"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/docs/examples"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
)

func TestImmutableAct(t *testing.T) {
	s := testcase.NewSpec(t)

	myStruct := s.Let(`MyStruct`, func(t *testcase.T) interface{} {
		return &examples.MyStruct{}
	})

	s.Describe(`#Shrug`, func(s *testcase.Spec) {
		const shrugEmoji = `¯\_(ツ)_/¯`
		var (
			message    = s.LetValue(`shrug message`, fixtures.Random.String())
			messageGet = func(t *testcase.T) string { return message.Get(t).(string) }
			subject    = func(t *testcase.T) string {
				return myStruct.Get(t).(*examples.MyStruct).Shrug(messageGet(t))
			}
		)

		s.When(`message doesn't have shrug in the ending`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.False(t, strings.HasSuffix(messageGet(t), shrugEmoji))
			})

			s.Then(`it will append shrug emoji to this`, func(t *testcase.T) {
				require.Equal(t, messageGet(t)+` `+shrugEmoji, subject(t))
			})
		})

		s.When(`shrug part of the input message`, func(s *testcase.Spec) {
			message.LetValue(s, fixtures.Random.String()+shrugEmoji)

			s.Then(`it will not append any more shrug emoji to the end of the message`, func(t *testcase.T) {
				require.Equal(t, messageGet(t), subject(t))
			})
		})
	})
}
