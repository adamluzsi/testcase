package testcase_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal/example/mydomain"
	"github.com/adamluzsi/testcase/internal/example/storages"
)

// package spechelper

var (
	ExampleStorage = testcase.Var[*storages.Storage]{
		ID: "storage component (external resource supplier)",
		Init: func(t *testcase.T) *storages.Storage {
			storage, err := storages.New(os.Getenv(`TEST_DATABASE_URL`))
			t.Must.Nil(err)
			t.Defer(storage.Close)
			return storage
		},
	}
	ExampleStorageGet = func(t *testcase.T) *storages.Storage {
		// workaround until go type parameter release
		return ExampleStorage.Get(t)
	}
	ExampleMyDomainUseCase = testcase.Var[*mydomain.MyUseCaseInteractor]{
		ID: "my domain rule (domain interactor)",
		Init: func(t *testcase.T) *mydomain.MyUseCaseInteractor {
			return &mydomain.MyUseCaseInteractor{Storage: ExampleStorageGet(t)}
		},
	}
	ExampleMyDomainUseCaseGet = func(t *testcase.T) *mydomain.MyUseCaseInteractor {
		// workaround until go type parameter release
		return ExampleMyDomainUseCase.Get(t)
	}
)

// package httpapi // external interface

func NewAPI(interactor *mydomain.MyUseCaseInteractor) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(`/foo`, func(w http.ResponseWriter, r *http.Request) {
		reply, err := interactor.Foo(r.Context())
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}
		_, _ = fmt.Fprint(w, reply)
	})
	return mux
}

// package httpapi_test

func ExampleVar_spechelper() {
	var t *testing.T
	s := testcase.NewSpec(t)

	api := testcase.Let(s, func(t *testcase.T) *http.ServeMux {
		return NewAPI(ExampleMyDomainUseCaseGet(t))
	})

	s.Describe(`GET /foo`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) *httptest.ResponseRecorder {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, `/`, nil)
			api.Get(t).ServeHTTP(w, r)
			return w
		}

		s.Then(`it will reply with baz`, func(t *testcase.T) {
			t.Must.Contain(`baz`, subject(t).Body.String())
		})
	})
}
