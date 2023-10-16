package testcase_test

import (
	"sync"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
)

func TestTableTest_orderIsDeterministic(t *testing.T) {
	t.Setenv(testcase.EnvKeySeed, "42")

	var cases = map[string]int{
		"1": 1,
		"2": 2,
		"3": 3,
		"4": 4,
		"5": 5,
		"6": 6,
		"7": 7,
	}

	t.Run("TT+Spec", func(t *testing.T) {
		var out []int
		s := testcase.NewSpec(t)
		testcase.TableTest(s, cases, func(t *testcase.T, v int) {
			out = append(out, v)
		})
		assert.Equal(t, []int{2, 7, 5, 6, 4, 1, 3}, out)
	})
	t.Run("TT+TestingTB", func(t *testing.T) {
		var out []int
		testcase.TableTest(t, cases, func(t *testcase.T, v int) {
			out = append(out, v)
		})
		assert.Equal(t, []int{2, 7, 5, 6, 4, 1, 3}, out)
	})
}

func TestTableTest_forIterationWorksWellInParallel(t *testing.T) {
	var (
		m     sync.Mutex
		out   = map[int]struct{}{}
		wg    sync.WaitGroup
		cases = map[string]int{
			"1": 1,
			"2": 2,
			"3": 3,
		}
	)

	wg.Add(3)

	s := testcase.NewSpec(t)
	s.Parallel()
	testcase.TableTest(s, cases, func(t *testcase.T, v int) {
		defer wg.Done()
		m.Lock()
		defer m.Unlock()
		out[v] = struct{}{}
	})
	s.Test("", func(t *testcase.T) {
		wg.Wait()
		expected := map[int]struct{}{
			1: {},
			2: {},
			3: {},
		}
		t.Must.ContainExactly(expected, out)
	})
}

func TestTableTest_classic(t *testing.T) {
	testCases := map[string]int{
		"on 42": 42,
		"on 24": 24,
	}
	t.Run("each test is executed", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		var out []int
		testcase.TableTest(s, testCases, func(t *testcase.T, val int) {
			out = append(out, val)
		})
		s.Finish()

		assert.ContainExactly(t, []int{42, 24}, out)
	})
	t.Run("values from the Spec context is inherited", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		v := testcase.LetValue(s, 42)
		testcase.TableTest(s, testCases, func(t *testcase.T, tc int) {
			t.Must.Equal(42, v.Get(t))
		})
	})
	t.Run("each test run in isolation", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		v := testcase.LetValue(s, 42)
		testcase.TableTest(s, testCases, func(t *testcase.T, tc int) {
			t.Must.Equal(42, v.Get(t), "the other table test should have no side effect on this test")
			v.Set(t, 24)
		})
		s.Finish()
	})
}

func TestTableTest_withTestBlock(t *testing.T) {
	v := testcase.Var[int]{ID: "the int value"}
	testCases := map[string]func(t *testcase.T){
		"A": func(t *testcase.T) {
			v.Set(t, 1)
		},
		"B": func(t *testcase.T) {
			v.Set(t, 2)
		},
		"C": func(t *testcase.T) {
			v.Set(t, 3)
		},
	}
	t.Run("each test is executed", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		var out []int
		testcase.TableTest(s, testCases, func(t *testcase.T) {
			out = append(out, v.Get(t))
		})
		s.Finish()

		assert.ContainExactly(t, []int{1, 2, 3}, out)
	})
	t.Run("values from the Spec context is inherited", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		val := testcase.LetValue(s, 42)
		testcase.TableTest(s, testCases, func(t *testcase.T) {
			t.Must.Equal(42, val.Get(t))
		})
	})
	t.Run("each test run in isolation", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		val := testcase.LetValue(s, 42)
		testcase.TableTest(s, testCases, func(t *testcase.T) {
			t.Must.Equal(42, val.Get(t), "the other table test should have no side effect on this test")
			val.Set(t, 24)
		})
		s.Finish()
	})
}

func TestTableTest_withSpecBlock(t *testing.T) {
	v := testcase.Var[int]{ID: "the int value"}
	testCases := map[string]func(*testcase.Spec){
		"A": func(s *testcase.Spec) {
			v.LetValue(s, 1)
		},
		"B": func(s *testcase.Spec) {
			v.LetValue(s, 2)
		},
		"C": func(s *testcase.Spec) {
			v.LetValue(s, 3)
		},
	}
	t.Run("each test is executed", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		var out []int
		testcase.TableTest(s, testCases, func(t *testcase.T) {
			out = append(out, v.Get(t))
		})
		s.Finish()

		assert.ContainExactly(t, []int{1, 2, 3}, out)
	})
	t.Run("values from the Spec context is inherited", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		val := testcase.LetValue(s, 42)
		testcase.TableTest(s, testCases, func(t *testcase.T) {
			t.Must.Equal(42, val.Get(t))
		})
	})
	t.Run("each test run in isolation", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		val := testcase.LetValue(s, 42)
		testcase.TableTest(s, testCases, func(t *testcase.T) {
			t.Must.Equal(42, val.Get(t), "the other table test should have no side effect on this test")
			val.Set(t, 24)
		})
		s.Finish()
	})
}

func TestTableTest_actAsSpec(t *testing.T) {
	v := testcase.Var[int]{ID: "the int value"}
	testCases := map[string]func(*testcase.Spec){
		"A": func(s *testcase.Spec) {
			v.LetValue(s, 1)
		},
		"B": func(s *testcase.Spec) {
			v.LetValue(s, 2)
		},
		"C": func(s *testcase.Spec) {
			v.LetValue(s, 3)
		},
	}
	t.Run("each test is executed", func(t *testing.T) {
		var out []int
		s := testcase.NewSpec(t)
		s.HasSideEffect()
		testcase.TableTest(s, testCases, func(s *testcase.Spec) {
			s.Test("", func(t *testcase.T) {
				out = append(out, v.Get(t))
			})
		})
		s.Finish()

		assert.ContainExactly(t, []int{1, 2, 3}, out)
	})
	t.Run("values from the Spec context is inherited", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()
		val := testcase.LetValue(s, 42)
		testcase.TableTest(s, testCases, func(s *testcase.Spec) {
			s.Test("", func(t *testcase.T) {
				t.Must.Equal(42, val.Get(t))
			})
		})
	})
	t.Run("each test run in isolation", func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.HasSideEffect()

		val := testcase.LetValue(s, 42)
		testcase.TableTest(s, testCases, func(s *testcase.Spec) {
			s.Test("", func(t *testcase.T) {
				t.Must.Equal(42, val.Get(t), "the other table test should have no side effect on this test")
				val.Set(t, 24)
			})
		})
		s.Finish()
	})
	t.Run("when concrete value is used with spec block", func(t *testing.T) {
		assert.Panic(t, func() {
			testcase.TableTest(t, map[string]struct{}{"hello": {}}, func(s *testcase.Spec) {})
		})
	})
}
