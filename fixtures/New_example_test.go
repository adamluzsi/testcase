package fixtures_test

import "github.com/adamluzsi/testcase/fixtures"

func ExampleNew() {
	type ExampleStruct struct {
		Name string
	}

	_ = fixtures.New(ExampleStruct{}).(*ExampleStruct)
}
