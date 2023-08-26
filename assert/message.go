package assert

type Message string

func toMsg(msg []Message) []any {
	var out []any
	for _, m := range msg {
		out = append(out, m)
	}
	return out
}
