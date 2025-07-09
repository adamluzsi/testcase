package fmterror_test

import (
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/fmterror"
	"go.llib.dev/testcase/pp"
)

func TestMessage_String(t *testing.T) {
	type TestCase struct {
		Message  fmterror.Message
		Expected string
	}
	for _, tc := range []TestCase{
		{
			Message:  fmterror.Message{},
			Expected: "",
		},
		{
			Message: fmterror.Message{
				Name: "Test",
			},
			Expected: "[Test] ",
		},
		{
			Message: fmterror.Message{
				Name:  "Test",
				Cause: "This",
			},
			Expected: "[Test] This",
		},
		{
			Message: fmterror.Message{
				Name:    "Test",
				Cause:   "This",
				Message: []interface{}{"out", 42},
			},
			Expected: "[Test] This\nout 42",
		},
		{
			Message: fmterror.Message{
				Name:  "Test",
				Cause: "This",
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
				Name:  "Test",
				Cause: "This",
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
		{
			Message: fmterror.Message{
				Values: []fmterror.Value{
					{
						Label: "foo",
						Value: fmterror.Formatted("hello"),
					},
				},
			},
			Expected: "\nfoo:\t" + "hello",
		},
	} {
		tc := tc
		t.Run(``, func(t *testing.T) {
			actual := tc.Message.String()
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
