package pp

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestFPP(t *testing.T) {
	buf := &bytes.Buffer{}
	v1 := time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)
	v2 := "foo"
	n, err := FPP(buf, v1, v2)
	if err != nil {
		t.Fatal(err.Error())
	}

	exp := "time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)\t\"foo\"\n"
	if len([]byte(exp)) != n {
		t.Fatal("not everything was written out")
	}

	act := buf.String()

	mustEqual(t, exp, act)
}

func TestPP(t *testing.T) {
	ogw := defaultWriter
	defer func() { defaultWriter = ogw }()

	buf := &bytes.Buffer{}
	defaultWriter = buf

	v1 := time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)
	v2 := "bar"

	_, file, line, _ := runtime.Caller(0)
	PP(v1, v2)

	exp := fmt.Sprintf("%s:%d time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)\t\"bar\"\n",
		filepath.Base(file), line+1)
	act := buf.String()

	mustEqual(t, exp, act)
}
