package assert

import "fmt"

type Message string

func MessageF(format string, args ...any) Message {
	return Message(fmt.Sprintf(format, args...))
}

func toMsg(msg []Message) []any {
	var out []any
	for _, m := range msg {
		out = append(out, m)
	}
	return out
}
