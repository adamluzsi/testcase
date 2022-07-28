package pp

import (
	"fmt"
	"os"
	"strconv"
)

var debug = false

func init() {
	v, ok := os.LookupEnv("TESTCASE_PP_DEBUG")
	if !ok {
		return
	}
	state, err := strconv.ParseBool(v)
	if err != nil {
		panic(err.Error())
	}
	debug = state
}

func debugRecover() {
	r := recover()
	if r == nil {
		return
	}
	if !debug {
		return
	}
	fmt.Println(r)
}
