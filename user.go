package testcase

import "os"

var userEnvVars = []string{
	`TESTCASE_USER`,
	`USER`,
	`USERNAME`,
	`LOGNAME`,
}

func getUser() string {
	for _, key := range userEnvVars {
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
	}
	return ""
}
