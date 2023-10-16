package testcase_test

import (
	"testing"
	"time"

	"go.llib.dev/testcase"
)

func TestOutput(t *testing.T) {
	if !testing.Verbose() {
		t.Skip()
	}

	s := testcase.NewSpec(t)

	s.Describe(`#A`, func(s *testcase.Spec) {
		s.Test(`foo`, func(t *testcase.T) {})
		s.Test(`bar`, func(t *testcase.T) {})
		s.Test(`baz`, func(t *testcase.T) {})
	})

	s.Describe(`#B`, func(s *testcase.Spec) {
		s.Test(`foo`, func(t *testcase.T) {})
		s.Test(`bar`, func(t *testcase.T) {})
		s.Test(`baz`, func(t *testcase.T) {})
	})

	testcase.RunSuite(s, OutputExampleContract{})
	testcase.RunOpenSuite(s, OutputExampleOpenContract{})

	s.Describe(`name-escapes`, func(s *testcase.Spec) {
		s.Test(`.`, func(t *testcase.T) {})
		s.Test(`+`, func(t *testcase.T) {})
		s.Test(`"`, func(t *testcase.T) {})
		s.Test(`'`, func(t *testcase.T) {})
		s.Test(`_`, func(t *testcase.T) {})
		s.Test(` `, func(t *testcase.T) {})
		s.Test(`,`, func(t *testcase.T) {})
		s.Test(`;`, func(t *testcase.T) {})
		s.Test(`+[].?`, func(t *testcase.T) {})
		s.Describe(`${PATH}`, func(s *testcase.Spec) {
			s.Test(``, func(t *testcase.T) {})
		})
	})
}

func BenchmarkOutput(b *testing.B) {
	s := testcase.NewSpec(b)

	s.Describe(`#A`, func(s *testcase.Spec) {
		s.Test(`foo`, func(t *testcase.T) { time.Sleep(time.Millisecond) })
		s.Test(`bar`, func(t *testcase.T) { time.Sleep(time.Millisecond) })
		s.Test(`baz`, func(t *testcase.T) { time.Sleep(time.Millisecond) })
	})

	s.Describe(`#B`, func(s *testcase.Spec) {
		s.Test(`foo`, func(t *testcase.T) { time.Sleep(time.Millisecond) })
		s.Test(`bar`, func(t *testcase.T) { time.Sleep(time.Millisecond) })
		s.Test(`baz`, func(t *testcase.T) { time.Sleep(time.Millisecond) })
	})
}

func TestComplexOutput(t *testing.T) {
	if !testing.Verbose() {
		t.Skip()
	}

	s := testcase.NewSpec(t)
	s.Describe(`1`, func(s *testcase.Spec) {
		s.When(`2`, func(s *testcase.Spec) {
			s.Then(`3`, func(t *testcase.T) {
				t.TB.(*testing.T).Run(`Run`, func(t *testing.T) {
					s := testcase.NewSpec(t)
					s.Describe(`4`, func(s *testcase.Spec) {
						s.When(`5`, func(s *testcase.Spec) {
							s.Then(`6`, func(t *testcase.T) {
								t.TB.(*testing.T).Run(`Run`, func(t *testing.T) {
									s := testcase.NewSpec(t)
									s.Describe(`7`, func(s *testcase.Spec) {
										s.When(`8`, func(s *testcase.Spec) {
											s.Then(`9`, func(t *testcase.T) {
												t.TB.(*testing.T).Run(`Run`, func(t *testing.T) {
													s := testcase.NewSpec(t)
													s.Describe(`10`, func(s *testcase.Spec) {
														s.When(`11`, func(s *testcase.Spec) {
															s.Then(`12`, func(t *testcase.T) {
																t.Log(`done`)
															})
														})
													})
												})
											})
										})
									})
								})
							})
						})
					})
				})
			})
		})
	})

	s = testcase.NewSpec(t)
	s.Describe(`subject`, func(s *testcase.Spec) {
		s.When(`spec - A`, func(s *testcase.Spec) {
			s.And(`spec - A`, func(s *testcase.Spec) {
				s.Then(`testCase`, func(t *testcase.T) {})
			})

			s.And(`spec - B`, func(s *testcase.Spec) {
				s.Then(`testCase`, func(t *testcase.T) {})
			})
		})

		s.When(`spec - B`, func(s *testcase.Spec) {
			s.And(`spec - A`, func(s *testcase.Spec) {
				s.Then(`testCase`, func(t *testcase.T) {})
			})

			s.And(`spec - B`, func(s *testcase.Spec) {
				s.Then(`testCase`, func(t *testcase.T) {})
			})
		})
	})
}

type OutputExampleOpenContract struct{}

func (c OutputExampleOpenContract) Test(t *testing.T) {
	t.Log(`OutputExampleOpenContract.Test`)
}

func (c OutputExampleOpenContract) Benchmark(b *testing.B) {
	b.Log(`OutputExampleOpenContract.Benchmark`)
}

type OutputExampleContract struct{}

func (c OutputExampleContract) Spec(s *testcase.Spec) {
	s.Test(``, func(t *testcase.T) {
		t.Log("OutputExampleOpenContract.Spec")
	})
}
