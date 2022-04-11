package fixtures_test

import "github.com/adamluzsi/testcase/fixtures"

func ExampleNew() {
	type ExampleStruct struct {
		Name string
	}

	var _ *ExampleStruct = fixtures.New[ExampleStruct]()
}
