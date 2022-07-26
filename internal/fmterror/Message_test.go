package fmterror_test

import (
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/fmterror"
	"github.com/adamluzsi/testcase/pp"
)

func TestMessage_String(t *testing.T) {
	type TestCase struct {
		Message  fmterror.Message
		Expected string
	}
	for _, tc := range []TestCase{
		{
			Message: fmterror.Message{
				Method: "Test",
			},
			Expected: "[Test] ",
		},
		{
			Message: fmterror.Message{
				Method: "Test",
				Cause:  "This",
			},
			Expected: "[Test] This",
		},
		{
			Message: fmterror.Message{
				Method:  "Test",
				Cause:   "This",
				Message: []interface{}{"out", 42},
			},
			Expected: "[Test] This\nout 42",
		},
		{
			Message: fmterror.Message{
				Method: "Test",
				Cause:  "This",
				Values: []fmterror.Value{
					{
						Label: "left-label",
						Value: 42,
					},
				},
				Message: []interface{}{"out", 42},
			},
			Expected: "[Test] This\nout 42\nleft-label:\t42",
		},
		{
			Message: fmterror.Message{
				Method: "Test",
				Cause:  "This",
				Values: []fmterror.Value{
					{
						Label: "left-label",
						Value: 42,
					},
					{
						Label: "right-label",
						Value: 24,
					},
				},
				Message: []interface{}{"out", 42},
			},
			Expected: "[Test] This\nout 42\n left-label:\t42\nright-label:\t24",
		},
		{
			Message: fmterror.Message{
				Values: []fmterror.Value{
					{
						Label: ".....",
						Value: 42,
					},
					{
						Label: "...",
						Value: 24,
					},
				},
			},
			Expected: "\n.....:\t42\n  ...:\t24",
		},
		{
			Message: fmterror.Message{
				Values: []fmterror.Value{
					{
						Label: "...",
						Value: 42,
					},
					{
						Label: ".....",
						Value: 24,
					},
				},
			},
			Expected: "\n  ...:\t42\n.....:\t24",
		},
		{
			Message: fmterror.Message{
				Values: []fmterror.Value{
					{
						Label: "foo",
						Value: []int{1, 2, 3},
					},
				},
			},
			Expected: "\nfoo:\n\n" + pp.Format([]int{1, 2, 3}) + "\n",
		},
	} {
		tc := tc
		t.Run(``, func(t *testing.T) {
			actual := tc.Message.String()
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
