//go:generate bash ./generate.sh
package mocks

import (
	"os"
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase/internal"
	"github.com/golang/mock/gomock"
)

func New(tb testing.TB, expectations ...func(mock *MockTB)) (*MockTB, func()) {
	ctrl := gomock.NewController(tb)
	tb.Cleanup(ctrl.Finish)
	mock := NewMockTB(ctrl)
	for _, expectation := range expectations {
		expectation(mock)
	}
	cleanupNow := SetupDefaultBehavior(tb, mock)
	return mock, cleanupNow
}

func NewMock(tb testing.TB, expectations ...func(mock *MockTB)) *MockTB {
	m, _ := New(tb, expectations...)
	return m
}

// DEPRECATED
func NewWithDefaults(tb testing.TB, expectations func(mock *MockTB)) *MockTB {
	return NewMock(tb, expectations)
}

func SetupDefaultBehavior(tb testing.TB, mock *MockTB) func() {
	mock.EXPECT().Helper().AnyTimes()
	mock.EXPECT().Log(gomock.Any()).AnyTimes()
	mock.EXPECT().Logf(gomock.Any(), gomock.Any()).AnyTimes()
	mock.EXPECT().TempDir().Return(os.TempDir()).AnyTimes()
	mock.EXPECT().Helper().AnyTimes()
	mock.EXPECT().Name().Return(tb.Name()).AnyTimes()
	mock.EXPECT().Run(gomock.Any(), gomock.Any()).Do(func(_ string, blk func(tb testing.TB)) bool {
		sub, td := New(tb, func(*MockTB) {})
		internal.InGoroutine(func() {
			defer td()
			blk(sub)
		})
		if sub.Failed() {
			mock.Fail()
		}
		return sub.Failed()
	}).Return(true).AnyTimes()

	var failed bool
	mock.EXPECT().Failed().DoAndReturn(func() bool { return failed }).AnyTimes()
	mock.EXPECT().Fail().Do(func() { failed = true }).AnyTimes()
	mock.EXPECT().FailNow().Do(func() { mock.Fail(); runtime.Goexit() }).AnyTimes()
	mock.EXPECT().Error(gomock.Any()).Do(func(...interface{}) { mock.Fail() }).AnyTimes()
	mock.EXPECT().Errorf(gomock.Any(), gomock.Any()).Do(func(string, ...interface{}) { mock.Fail() }).AnyTimes()
	mock.EXPECT().Fatal(gomock.Any()).Do(func(...interface{}) { mock.FailNow() }).AnyTimes()
	mock.EXPECT().Fatalf(gomock.Any(), gomock.Any()).Do(func(string, ...interface{}) { mock.FailNow() }).AnyTimes()

	var skipped bool
	mock.EXPECT().Skipped().DoAndReturn(func() bool { return skipped }).AnyTimes()
	mock.EXPECT().Skip(gomock.Any()).Do(func(...interface{}) { mock.SkipNow() }).AnyTimes()
	mock.EXPECT().Skipf(gomock.Any(), gomock.Any()).Do(func(string, ...interface{}) { mock.SkipNow() }).AnyTimes()
	mock.EXPECT().SkipNow().Do(func() { skipped = true; runtime.Goexit() }).AnyTimes()

	cleanups := make([]func(), 0)
	mock.EXPECT().Cleanup(gomock.Any()).Do(func(fn func()) { cleanups = append(cleanups, fn) }).AnyTimes()

	return func() {
		for _, c := range cleanups {
			defer internal.InGoroutine(c)
		}
	}
}
