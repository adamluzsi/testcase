package testcase

type visitor interface {
	Visit(s *Spec)
}

type visitable interface {
	acceptVisitor(visitor)
}

type visitorFunc func(s *Spec)

func (fn visitorFunc) Visit(s *Spec) { fn(s) }
