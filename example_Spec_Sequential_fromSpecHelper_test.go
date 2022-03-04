package testcase_test

import (
	"context"
	"os"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal/example/mydomain"
)

func ExampleSpec_Sequential_fromSpecHelper() {
	var t *testing.T
	s := testcase.NewSpec(t)
	Setup(s) // setup specification with spec helper function

	// Tells that the subject of this specification should be software side effect free on its own.
	s.NoSideEffect()

	var (
		myUseCase = func(t *testcase.T) *MyUseCaseThatHasStorageDependency {
			return &MyUseCaseThatHasStorageDependency{Storage: Storage.Get(t)}
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

var Storage = testcase.Var[mydomain.Storage]{ID: `storage`}

func Setup(s *testcase.Spec) {
	// spec helper function that is environment aware, and can decide what resource should be used in the testCase runtime.
	env, ok := os.LookupEnv(`TEST_DB_CONNECTION_URL`)

	if ok {
		s.Sequential()
		// or
		s.HasSideEffect()
		Storage.Let(s, func(t *testcase.T) mydomain.Storage {
			// open database connection
			_ = env // use env to connect or something
			// setup isolation with tx
			return &ExternalResourceBasedStorage{ /*...*/ }
		})
	} else {
		Storage.Let(s, func(t *testcase.T) mydomain.Storage {
			return &InMemoryBasedStorage{}
		})
	}
}

type InMemoryBasedStorage struct{}

func (i InMemoryBasedStorage) BeginTx(ctx context.Context) (context.Context, error) { panic("") }
func (i InMemoryBasedStorage) CommitTx(ctx context.Context) error                   { panic("") }
func (i InMemoryBasedStorage) RollbackTx(ctx context.Context) error                 { panic("") }

type ExternalResourceBasedStorage struct{}

func (e ExternalResourceBasedStorage) BeginTx(ctx context.Context) (context.Context, error) {
	panic("")
}
func (e ExternalResourceBasedStorage) CommitTx(ctx context.Context) error   { panic("") }
func (e ExternalResourceBasedStorage) RollbackTx(ctx context.Context) error { panic("") }

type MyUseCaseThatHasStorageDependency struct {
	Storage MyUseCaseStorageRoleInterface
}

func (d *MyUseCaseThatHasStorageDependency) SomeMethod() bool {
	return false
}

type MyUseCaseStorageRoleInterface interface{}
