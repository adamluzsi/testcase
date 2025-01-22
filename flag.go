package testcase

import (
	"os"
	"strconv"
	"testing"
)

func EnvFlag( /* const */ EnvVarName string, opts ...EnvFlagOpt) func(testing.TB) {
	if len(EnvVarName) == 0 {
		panic("testcase.EnvFlag must be used with a non empty environment variable name")
	}
	var c configEnvFlag
	for _, opt := range opts {
		opt.configure(&c)
	}
	return func(tb testing.TB) {
		flag, ok := os.LookupEnv(EnvVarName)
		if !ok {
			tb.Logf("missing environment variable: %s", EnvVarName)
			if c.Required {
				tb.FailNow()
			} else {
				tb.SkipNow()
			}
		}

		enabled, err := strconv.ParseBool(flag)
		if err != nil {
			tb.Logf("failed to parse %s flag env var: %s", EnvVarName, err.Error())
			tb.FailNow()
		}

		if !enabled {
			tb.Logf("[SKIP] %s flagged as disabled", EnvVarName)
			tb.SkipNow()
		}
	}
}

type EnvFlagOpt interface{ configure(*configEnvFlag) }

func EnvFlagRequired() EnvFlagOpt {
	return envFlagOptFunc(func(cef *configEnvFlag) {
		cef.Required = true
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type envFlagOptFunc func(*configEnvFlag)

func (fn envFlagOptFunc) configure(c *configEnvFlag) { fn(c) }

type configEnvFlag struct {
	Required bool
}
