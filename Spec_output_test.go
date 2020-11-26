package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func TestOutput_short(t *testing.T) {
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
}

func TestOutput(t *testing.T) {
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
		s.When(`context - A`, func(s *testcase.Spec) {
			s.And(`context - A`, func(s *testcase.Spec) {
				s.Then(`test`, func(t *testcase.T) {})
			})

			s.And(`context - B`, func(s *testcase.Spec) {
				s.Then(`test`, func(t *testcase.T) {})
			})
		})

		s.When(`context - B`, func(s *testcase.Spec) {
			s.And(`context - A`, func(s *testcase.Spec) {
				s.Then(`test`, func(t *testcase.T) {})
			})

			s.And(`context - B`, func(s *testcase.Spec) {
				s.Then(`test`, func(t *testcase.T) {})
			})
		})
	})

}
