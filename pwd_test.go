package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/sandbox"
	"os"
	"path/filepath"
	"testing"
)

func TestChdir(t *testing.T) {
	s := testcase.NewSpec(t)

	PWD, err := os.Getwd()
	assert.NoError(t, err)

	var (
		tb = testcase.Let(s, func(t *testcase.T) *doubles.TB {
			dtb := &doubles.TB{}
			t.Cleanup(dtb.Finish)
			return dtb
		})
		dir = testcase.Let[string](s, nil)
	)
	act := func(t *testcase.T) {
		testcase.Chdir(tb.Get(t), dir.Get(t))
	}

	s.When("dir points to a valid directory path", func(s *testcase.Spec) {
		dir.LetValue(s, "./internal")

		s.Then("it runs without an issue", func(t *testcase.T) {
			act(t)
			t.Must.False(tb.Get(t).Failed())
		})

		s.Then("it will change working directory", func(t *testcase.T) {
			act(t)

			currentPWD, err := os.Getwd()
			t.Must.NoError(err)
			t.Must.NotEqual(PWD, currentPWD)
			t.Must.Equal(currentPWD, filepath.Join(PWD, dir.Get(t)))
		})

		s.Then("on testing.TB.Cleanup the directory is restored", func(t *testcase.T) {
			PWD, err := os.Getwd()
			t.Must.NoError(err)

			act(t)
			tb.Get(t).Finish() // after cleanup

			currentPWD, err := os.Getwd()
			t.Must.NoError(err)
			t.Must.Equal(PWD, currentPWD)
		})

		s.When("chdir was already called in the given test", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testcase.Chdir(tb.Get(t), ".")
			})

			s.Then("it still able to change to the given directory", func(t *testcase.T) {
				act(t)

				currentPWD, err := os.Getwd()
				t.Must.NoError(err)
				t.Must.NotEqual(PWD, currentPWD)
				t.Must.Equal(currentPWD, filepath.Join(PWD, dir.Get(t)))
			})
		})
	})

	s.When("director is invalid", func(s *testcase.Spec) {
		dir.LetValue(s, "./invalid-directory-path")

		s.Then("the test will fail", func(t *testcase.T) {
			sandbox.Run(func() {
				act(t)
			})

			t.Must.True(tb.Get(t).Failed())
		})
	})
}

func TestChdir_withParallel(t *testing.T) {
	PWD, err := os.Getwd()
	assert.NoError(t, err)

	t.Run("foo", func(t *testing.T) {
		t.Parallel()
		testcase.Chdir(t, "./internal")

		currentPWD, err := os.Getwd()
		assert.NoError(t, err)
		assert.NotEqual(t, PWD, currentPWD)
		assert.Equal(t, currentPWD, filepath.Join(PWD, "internal"))
	})
	t.Run("bar", func(t *testing.T) {
		t.Parallel()
		testcase.Chdir(t, "./random")

		currentPWD, err := os.Getwd()
		assert.NoError(t, err)
		assert.NotEqual(t, PWD, currentPWD)
		assert.Equal(t, currentPWD, filepath.Join(PWD, "random"))
	})
	t.Run("baz", func(t *testing.T) {
		t.Parallel()
		testcase.Chdir(t, "./assert")

		currentPWD, err := os.Getwd()
		assert.NoError(t, err)
		assert.NotEqual(t, PWD, currentPWD)
		assert.Equal(t, currentPWD, filepath.Join(PWD, "assert"))
	})
}

//func TestChdir_willChangeDir(t *testing.T) {
//	PWD, err := os.Getwd()
//	assert.NoError(t, err)
//	testcase.Chdir(t, "./internal")
//	currentPWD, err := os.Getwd()
//	assert.NoError(t, err)
//	assert.NotEqual(t, PWD, currentPWD)
//	assert.Equal(t, currentPWD, filepath.Join(PWD, "internal"))
//}
