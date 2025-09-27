package assert_test

type Greeter interface{ Greet() }

type Foo struct{}

func (foo Foo) Greet() {}

type Bar struct{}

func (bar Bar) Greet() {}
