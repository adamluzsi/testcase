package faultinject

import (
	"os"
	"strings"
)

var Enabled bool

func init() { initEnabled() }

func initEnabled() {
	Enabled = true
	const envKey = "TESTCASE_FAULT_INJECTION"
	if v, ok := os.LookupEnv(envKey); ok {
		switch strings.ToUpper(v) {
		case "TRUE", "ON":
			Enabled = true
		case "FALSE", "OFF":
			Enabled = false
		}
	}
}

type testingTB interface {
	Cleanup(func())
}

func ForTest(tb testingTB, enabled bool) {
	og := Enabled
	tb.Cleanup(func() { Enabled = og })
	Enabled = enabled
}
