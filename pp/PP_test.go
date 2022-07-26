package pp

import (
	"bytes"
	"testing"
	"time"
)

func TestFPP(t *testing.T) {
	buf := &bytes.Buffer{}
	v1 := time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)
	v2 := "foo"
	FPP(buf, v1, v2)

	exp := "time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)\n\"foo\"\n"
	act := buf.String()
	if act != exp {
		t.Fatalf("exp:\n%s\n\nact:\n%s", exp, act)
	}
}

func TestPP(t *testing.T) {
	ogw := defaultWriter
	defer func() { defaultWriter = ogw }()

	buf := &bytes.Buffer{}
	defaultWriter = buf

	v1 := time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)
	v2 := "bar"
	PP(v1, v2)

	exp := "time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)\n\"bar\"\n"
	act := buf.String()
	if act != exp {
		t.Fatalf("exp:\n%s\n\nact:\n%s", exp, act)
	}
}
