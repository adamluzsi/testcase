package spechelper

import (
	"os"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/internal/example/memory"
	"go.llib.dev/testcase/internal/example/mydomain"
	"go.llib.dev/testcase/internal/example/someextres"
)

var Storage = testcase.Var[mydomain.Storage]{
	ID: `storage`,
	OnLet: func(s *testcase.Spec, v testcase.Var[mydomain.Storage]) {
		// spec helper function that is environment aware, and can decide what resource should be used in the testCase runtime.
		connstr, ok := os.LookupEnv(`TEST_DB_CONNECTION_URL`)

		if !ok {
			s.NoSideEffect()

			v.Let(s, func(t *testcase.T) mydomain.Storage {
				storage, err := someextres.NewStorage(connstr)
				t.Must.NoError(err)
				return storage
			})
			return
		}

		s.HasSideEffect()
		// or
		s.Sequential()

		v.Let(s, func(t *testcase.T) mydomain.Storage {
			return memory.NewStorage()
		})
	},
}

var (
	ExampleStorage = testcase.Var[*someextres.Storage]{
		ID: "storage component (external resource supplier)",
		Init: func(t *testcase.T) *someextres.Storage {
			storage, err := someextres.NewStorage(os.Getenv(`TEST_DATABASE_URL`))
			t.Must.Nil(err)
			t.Defer(storage.Close)
			return storage
		},
	}
	ExampleStorageGet = func(t *testcase.T) *someextres.Storage {
		// workaround until go type parameter release
		return ExampleStorage.Get(t)
	}
	ExampleMyDomainUseCase = testcase.Var[*mydomain.MyUseCase]{
		ID: "my domain rule (domain interactor)",
		Init: func(t *testcase.T) *mydomain.MyUseCase {
			return &mydomain.MyUseCase{Storage: ExampleStorageGet(t)}
		},
	}
	ExampleMyDomainUseCaseGet = func(t *testcase.T) *mydomain.MyUseCase {
		// workaround until go type parameter release
		return ExampleMyDomainUseCase.Get(t)
	}
)
