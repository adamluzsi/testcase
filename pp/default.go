package pp

import (
	"io"
	"os"
)

func init() {
	initDefaultWriter()
}

var defaultWriter io.Writer = os.Stderr

func initDefaultWriter() {
	fpath, ok := os.LookupEnv("PP")
	if !ok {
		return
	}
	if fpath == "" {
		return
	}
	stat, err := os.Stat(fpath)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if stat != nil && stat.IsDir() {
		return
	}
	out, err := os.OpenFile(fpath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defaultWriter = out
}
