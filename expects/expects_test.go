package expects_test

import (
	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/testcase/expects"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/sandbox"
	"testing"
)

func Test_smoke(t *testing.T) {
	type TC struct {
		Expect func(tb testing.TB)
		Failed bool
	}

	testcase.TableTest(t, map[string]TC{
		"Equal - happy": {
			Expect: func(tb testing.TB) {
				Expect(tb, "42").To(Equal("42"))
			},
			Failed: false,
		},
		"Equal - happy - slice": {
			Expect: func(tb testing.TB) {
				Expect(tb, []string{"42"}).To(Equal([]string{"42"}))
			},
			Failed: false,
		},
		"Equal - fail": {
			Expect: func(tb testing.TB) {
				Expect(tb, 42).To(Equal(12))
			},
			Failed: true,
		},
		"Match - fail": {
			Expect: func(tb testing.TB) {
				Expect(tb, 42).To(Equal(12))
			},
			Failed: true,
		},
	}, func(t *testcase.T, tc TC) {
		dtb := &doubles.TB{}
		out := sandbox.Run(func() { tc.Expect(dtb) })
		t.Must.Equal(tc.Failed, dtb.IsFailed)
		t.Must.Equal(dtb.IsFailed, !out.OK)
	})
}
