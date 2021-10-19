package assert

import "testing"

func TestMessage_String(t *testing.T) {
	type TestCase struct {
		Message  message
		Expected string
	}
	for _, tc := range []TestCase{
		{
			Message: message{
				Method: "Test",
			},
			Expected: "[Test] ",
		},
		{
			Message: message{
				Method: "Test",
				Cause:  "This",
			},
			Expected: "[Test] This",
		},
		{
			Message: message{
				Method:      "Test",
				Cause:       "This",
				UserMessage: []interface{}{"out", 42},
			},
			Expected: "[Test] This\nout 42",
		},
		{
			Message: message{
				Method: "Test",
				Cause:  "This",
				Left: &messageValue{
					Label: "left-label",
					Value: 42,
				},
				UserMessage: []interface{}{"out", 42},
			},
			Expected: "[Test] This\nout 42\nleft-label:\t42",
		},
		{
			Message: message{
				Method: "Test",
				Cause:  "This",
				Left: &messageValue{
					Label: "left-label",
					Value: 42,
				},
				Right: &messageValue{
					Label: "right-label",
					Value: 24,
				},
				UserMessage: []interface{}{"out", 42},
			},
			Expected: "[Test] This\nout 42\n left-label:\t42\nright-label:\t24",
		},
		{
			Message: message{
				Left: &messageValue{
					Label: ".....",
					Value: 42,
				},
				Right: &messageValue{
					Label: "...",
					Value: 24,
				},
			},
			Expected: "\n.....:\t42\n  ...:\t24",
		},
		{
			Message: message{
				Left: &messageValue{
					Label: "...",
					Value: 42,
				},
				Right: &messageValue{
					Label: ".....",
					Value: 24,
				},
			},
			Expected: "\n  ...:\t42\n.....:\t24",
		},
	} {
		tc := tc
		t.Run(``, func(t *testing.T) {
			actual := tc.Message.String()
			Must(t).Equal(tc.Expected, actual)
		})
	}
}
