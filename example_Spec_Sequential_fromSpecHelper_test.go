package testcase_test

import (
	"os"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Sequential_fromSpecHelper() {
	var t *testing.T
	s := testcase.NewSpec(t)
	Setup(s) // setup specification with spec helper function

	// Tells that the subject of this specification should be software side effect free on its own.
	s.NoSideEffect()

	var (
		myUseCase = func(t *testcase.T) *MyUseCaseThatHasStorageDependency {
			return &MyUseCaseThatHasStorageDependency{Storage: Storage.Get(t).(MyUseCaseStorageRoleInterface)}
		}
	)

	s.Describe(`#SomeMethod`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) bool {
			return myUseCase(t).SomeMethod()
		}

		s.Test(`it is expected ...`, func(t *testcase.T) {
			if !subject(t) {
				t.Fatal(`assertion failed`)
			}
		})
	})
}

///////////////////////////////////////// in some package testing / spechelper /////////////////////////////////////////

var Storage = testcase.Var{Name: `storage`}

func Setup(s *testcase.Spec) {
	// spec helper function that is environment aware, and can decide what resource should be used in the test runtime.
	env, ok := os.LookupEnv(`TEST_DB_CONNECTION_URL`)

	if ok {
		s.Sequential()
		// or
		s.HasSideEffect()
		Storage.Let(s, func(t *testcase.T) interface{} {
			// open database connection
			_ = env // use env to connect or something
			// setup isolation with tx
			return &ExternalResourceBasedStorage{ /*...*/ }
		})
	} else {
		Storage.Let(s, func(t *testcase.T) interface{} {
			return &InMemoryBasedStorage{}
		})
	}
}

type InMemoryBasedStorage struct{}

type ExternalResourceBasedStorage struct{}

type MyUseCaseThatHasStorageDependency struct {
	Storage MyUseCaseStorageRoleInterface
}

func (d *MyUseCaseThatHasStorageDependency) SomeMethod() bool {
	return false
}

type MyUseCaseStorageRoleInterface interface{}
