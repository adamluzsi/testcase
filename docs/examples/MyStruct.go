package examples

type MyStruct struct{}

func (ms MyStruct) Say() string {
	return `Hello, World!`
}

func (ms MyStruct) Foo() string {
	return `Bar`
}
