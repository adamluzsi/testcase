package doc_test

import (
	"context"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/internal/doc"
)

func TestTestDocumentGenerator(t *testing.T) {

	t.Run("no error (no verbose) - colourless", func(t *testing.T) {
		internal.StubVerbose(t, func() bool { return false })

		docw := doc.DocumentFormat{}

		d, err := docw.MakeDocument(context.Background(), []doc.TestingCase{
			{
				ContextPath: []string{
					"TestTestDocumentGenerator",
					"smoke",
					"testA",
				},
				TestFailed: false,
			},
			{
				ContextPath: []string{
					"TestTestDocumentGenerator",
					"smoke",
					"testB",
				},
				TestFailed: false,
			},
		})
		assert.NoError(t, err)
		assert.Empty(t, d)
	})

	t.Run("no error (verbose) - colourless", func(t *testing.T) {
		testcase.SetEnv(t, "TERM", "dumb")
		internal.StubVerbose(t, func() bool { return true })

		docw := doc.DocumentFormat{}

		d, err := docw.MakeDocument(context.Background(), []doc.TestingCase{
			{
				ContextPath: []string{
					"TestTestDocumentGenerator",
					"smoke",
					"testA",
				},
				TestFailed: false,
			},
		})
		assert.NoError(t, err)

		exp := "TestTestDocumentGenerator\n  smoke\n    testA\n"
		assert.Equal(t, d, exp)
	})

	t.Run("many - colourless", func(t *testing.T) {
		testcase.SetEnv(t, "TERM", "dumb")

		docw := doc.DocumentFormat{}

		d, err := docw.MakeDocument(context.Background(), []doc.TestingCase{
			{
				ContextPath: []string{
					"TestTestDocumentGenerator",
					"smoke",
					"testA",
				},
				TestFailed: false,
			},
			{
				ContextPath: []string{
					"TestTestDocumentGenerator",
					"smoke",
					"testB",
				},
				TestFailed: true,
			},
		})
		assert.NoError(t, err)

		base := "TestTestDocumentGenerator\n"
		base += "  smoke\n"
		exp1 := base + "    testA\n" + "    testB [FAIL]\n"
		exp2 := base + "    testB [FAIL]\n" + "    testA\n"

		assert.AnyOf(t, func(a *assert.A) {
			a.Case(func(t assert.It) { assert.Contain(t, d, exp1) })
			a.Case(func(t assert.It) { assert.Contain(t, d, exp2) })
		})
	})

	t.Run("many - colourised", func(t *testing.T) {
		testcase.SetEnv(t, "TERM", "xterm-256color")

		docw := doc.DocumentFormat{}

		d, err := docw.MakeDocument(context.Background(), []doc.TestingCase{
			{
				ContextPath: []string{
					"TestTestDocumentGenerator",
					"smoke",
					"testA",
				},
				TestFailed: false,
			},
			{
				ContextPath: []string{
					"TestTestDocumentGenerator",
					"smoke",
					"testB",
				},
				TestFailed: true,
			},
		})
		assert.NoError(t, err)

		exp1 := "TestTestDocumentGenerator\n  smoke\n    \x1b[91mtestB [FAIL]\x1b[0m\n    \x1b[92mtestA\x1b[0m\n"
		exp2 := "TestTestDocumentGenerator\n  smoke\n    \x1b[92mtestA\x1b[0m\n    \x1b[91mtestB [FAIL]\x1b[0m\n"

		assert.AnyOf(t, func(a *assert.A) {
			a.Case(func(t assert.It) { assert.Contain(t, d, exp1) })
			a.Case(func(t assert.It) { assert.Contain(t, d, exp2) })
		})
	})
}

func Test_spike(t *testing.T) {

	docw := doc.DocumentFormat{}

	d, err := docw.MakeDocument(context.Background(), []doc.TestingCase{
		{
			ContextPath: []string{
				"subject",
				"when",
				"and",
				"then A",
			},
			TestFailed: false,
		},
		{
			ContextPath: []string{
				"subject",
				"when",
				"and",
				"then B",
			},
			TestFailed: true,
		},
	})
	assert.NoError(t, err)

	t.Log("\n\n" + d)

}
