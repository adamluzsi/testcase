package testent

type Fooer interface{ Foo() }

type Foo struct{ ID string }

var _ Fooer = Foo{}

func (Foo) Foo() {}

type Bazer interface{ Baz() }

type Baz struct{ ID string }

var _ Bazer = Baz{}

func (Baz) Baz() {}
