package httpspec_test

import (
	"fmt"
	"net/http"

	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/testcase/httpspec"
)

func Example_usageWithDotImport() {
	s := testcase.NewSpec(testingT)

	Handler.Let(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	s.Before(func(t *testcase.T) {
		t.Log(`given authentication header is set`)
		Header.Get(t).Set(`X-Auth-Token`, `token`)
	})

	s.Describe(`GET / - list of X`, func(s *testcase.Spec) {
		Method.LetValue(s, http.MethodGet)
		Path.LetValue(s, `/`)

		var onSuccess = func(t *testcase.T) ListResponse {
			rr := ServeHTTP(t)
			t.Must.Equal(http.StatusOK, rr.Code)
			// unmarshal the response from rr.body
			return ListResponse{}
		}

		s.And(`something is set in the query`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				Query.Get(t).Set(`something`, `value`)
			})

			s.Then(`it will react to it as`, func(t *testcase.T) {
				listResponse := onSuccess(t)
				// assert
				_ = listResponse
			})
		})

		s.Then(`it will return the list of resource`, func(t *testcase.T) {
			listResponse := onSuccess(t)
			// assert
			_ = listResponse
		})
	})

	s.Describe(`GET /{resourceID} - show X`, func(s *testcase.Spec) {
		Method.LetValue(s, http.MethodGet)
		Path.Let(s, func(t *testcase.T) string {
			return fmt.Sprintf(`/%s`, t.I(`resourceID`))
		})

		var onSuccess = func(t *testcase.T) ShowResponse {
			rr := ServeHTTP(t)
			t.Must.Equal(http.StatusOK, rr.Code)
			// unmarshal the response from rr.body
			return ShowResponse{}
		}

		s.Then(`it will return the resource 'show'' representation`, func(t *testcase.T) {
			showResponse := onSuccess(t)
			// assert
			_ = showResponse
		})
	})
}
