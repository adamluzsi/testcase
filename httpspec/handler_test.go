package httpspec_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func Test_handlerSpec(t *testing.T) {
	s := testcase.NewSpec(t)

	// the behavior of the httpspec is tested through creating side effects.
	// Using side effect in an actual API specification is discouraged.
	var (
		ctx    context.Context
		path   string
		method string
		query  url.Values
		header http.Header
		body   []byte
	)
	s.Before(func(t *testcase.T) {
		ctx = nil
		method = ``
		path = ``
		query = nil
		header = nil
		body = nil
	})
	httpspec.HandlerSpec(s, func(t *testcase.T) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx = r.Context()
			method = r.Method
			path = r.URL.Path
			query = r.URL.Query()
			header = r.Header
			bs, err := ioutil.ReadAll(r.Body)
			require.Nil(t, err)
			require.Nil(t, r.Body.Close())
			body = bs
			w.Header().Set(`Hello`, `World`)
			w.WriteHeader(http.StatusTeapot)
			_, err = fmt.Fprint(w, `Hello, World!`)
			require.Nil(t, err)
		})
	})

	s.Describe(`httpspec.ServeHTTP`, func(s *testcase.Spec) {
		s.Then(`it should return a response recorder with the API response`, func(t *testcase.T) {
			rr := httpspec.ServeHTTP(t)
			require.Equal(t, http.StatusTeapot, rr.Code)
			require.Equal(t, `World`, rr.Header().Get(`Hello`))
			require.Equal(t, `Hello, World!`, rr.Body.String())
		})
	})

	s.When(`context defined`, func(s *testcase.Spec) {
		var expected = context.WithValue(context.Background(), `key`, `value`)
		httpspec.LetContext(s, func(t *testcase.T) context.Context { return expected })

		s.And(`using context key-value is added with testcase.T#Let + httpspec.ContextVarName`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				ctx := t.I(httpspec.ContextVarName).(context.Context)
				ctx = context.WithValue(ctx, `foo`, `bar`)
				t.Let(httpspec.ContextVarName, ctx)
			})

			s.Then(`in this scope the key-values of the context will be updated`, func(t *testcase.T) {
				httpspec.ServeHTTP(t)

				require.Equal(t, `bar`, ctx.Value(`foo`).(string))
			})
		})

		s.Then(`the context will be passed for the request`, func(t *testcase.T) {
			t.Log(`this can be used to create API specs where value in context is part of the http.handler prerequisite`)
			httpspec.ServeHTTP(t)
			require.Equal(t, expected, ctx)
		})
	})

	s.When(`query populated during Spec#Before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			httpspec.Query(t).Set(`hello`, `world`)
			httpspec.Query(t).Add(`l`, `a`)
			httpspec.Query(t).Add(`l`, `b`)
			httpspec.Query(t).Add(`l`, `c`)
		})

		s.Then(`it will pass the query to the request`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			require.ElementsMatch(t, []string{`a`, `b`, `c`}, query[`l`])
			require.Equal(t, `world`, query.Get(`hello`))
		})
	})

	s.When(`header populated during Spec#Before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			httpspec.Header(t).Set(`Hello`, `world`)
			httpspec.Header(t).Add(`L`, `a`)
			httpspec.Header(t).Add(`L`, `b`)
			httpspec.Header(t).Add(`L`, `c`)
		})

		s.Then(`it will HandlerSpec the headers for the request`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			t.Log(header)
			require.ElementsMatch(t, []string{`a`, `b`, `c`}, header[`L`])
			require.Equal(t, `world`, header.Get(`Hello`))
		})
	})

	s.When(`path is not defined`, func(s *testcase.Spec) {
		s.Then(`it will use / as default`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			require.Equal(t, `/`, path)
		})
	})

	s.When(`path is defined with LetPath`, func(s *testcase.Spec) {
		httpspec.LetPath(s, func(t *testcase.T) string { return `/hello/world` })

		s.Then(`it will call request with the given path`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			require.Equal(t, `/hello/world`, path)
		})
	})

	s.When(`path is defined with LetPathValue`, func(s *testcase.Spec) {
		httpspec.LetPathValue(s, `/foo/baz`)

		s.Then(`it will call request with the given path`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			require.Equal(t, `/foo/baz`, path)
		})
	})

	s.When(`method is not defined`, func(s *testcase.Spec) {
		s.Then(`it will use HTTP GET as default`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			require.Equal(t, http.MethodGet, method)
		})
	})

	s.When(`method is defined with LetMethod`, func(s *testcase.Spec) {
		httpspec.LetMethod(s, func(t *testcase.T) string { return http.MethodPost })

		s.Then(`it will use the http method for the request`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			require.Equal(t, http.MethodPost, method)
		})
	})

	s.When(`method is defined with LetMethodValue`, func(s *testcase.Spec) {
		httpspec.LetMethodValue(s, http.MethodPut)

		s.Then(`it will use the http method for the request`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			require.Equal(t, http.MethodPut, method)
		})
	})

	s.When(`body defined`, func(s *testcase.Spec) {
		const expected = `Hello, World!`

		s.Context(`as io.Reader`, func(s *testcase.Spec) {
			httpspec.LetBody(s, func(t *testcase.T) interface{} {
				return strings.NewReader(`Hello, World!`)
			})

			s.Then(`value is passed as is, without any further action`, func(t *testcase.T) {
				httpspec.ServeHTTP(t)
				actual := string(body)
				require.Equal(t, len(expected), len(actual))
				require.Equal(t, expected, actual)
			})

			s.And(`if debugging enabled`, func(s *testcase.Spec) {
				httpspec.Debug(s)

				s.Then(`it will pass the io reader content`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)
					actual := string(body)
					require.Equal(t, len(expected), len(actual))
					require.Equal(t, expected, actual)
				})
			})
		})

		s.Context(`as struct`, func(s *testcase.Spec) {
			s.And(`it has tags for form and json to define the keys`, func(s *testcase.Spec) {
				httpspec.LetBody(s, func(t *testcase.T) interface{} {
					return struct {
						Hello string `json:"hello_json_key" form:"hello_form_key"`
					}{Hello: `world`}
				})

				s.And(`form encoding is used`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
					})

					s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)

						require.Equal(t, `hello_form_key=world`, string(body))
					})
				})

				s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header(t).Set(`Content-Type`, `application/json`)
					})

					s.Then(`it will use json encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)
						expected := `{"hello_json_key":"world"}` + "\n"
						actual := string(body)
						require.Equal(t, len(expected), len(actual))
						require.Equal(t, expected, actual)
					})
				})
			})

			s.And(`it has no tags`, func(s *testcase.Spec) {
				httpspec.LetBody(s, func(t *testcase.T) interface{} {
					return struct{ TheKey string }{TheKey: `TheValue`}
				})

				s.And(`form encoding is used`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
					})

					s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)

						require.Equal(t, `TheKey=TheValue`, string(body))
					})
				})

				s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header(t).Set(`Content-Type`, `application/json`)
					})

					s.Then(`it will use json encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)
						expected := `{"TheKey":"TheValue"}` + "\n"
						actual := string(body)
						require.Equal(t, len(expected), len(actual))
						require.Equal(t, expected, actual)
					})
				})
			})
		})

		s.Context(`as pointer`, func(s *testcase.Spec) {
			s.Context(`to struct`, func(s *testcase.Spec) {
				httpspec.LetBody(s, func(t *testcase.T) interface{} {
					return &struct {
						Hello string `json:"hello_json" form:"hello_form"`
					}{Hello: `world`}
				})

				s.And(`form encoding is used`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
					})

					s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)

						require.Equal(t, `hello_form=world`, string(body))
					})
				})

				s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header(t).Set(`Content-Type`, `application/json`)
					})

					s.Then(`it will use json encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)
						expected := `{"hello_json":"world"}` + "\n"
						actual := string(body)
						require.Equal(t, len(expected), len(actual))
						require.Equal(t, expected, actual)
					})
				})
			})
		})

		s.Context(`as map string to string`, func(s *testcase.Spec) {
			httpspec.LetBody(s, func(t *testcase.T) interface{} {
				return map[string]string{"hello": "world"}
			})

			s.And(`form encoding is used`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
				})

				s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)

					require.Equal(t, `hello=world`, string(body))
				})
			})

			s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header(t).Set(`Content-Type`, `application/json`)
				})

				s.Then(`it will use json encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)
					expected := `{"hello":"world"}` + "\n"
					actual := string(body)
					require.Equal(t, len(expected), len(actual))
					require.Equal(t, expected, actual)
				})
			})
		})

		s.Context(`as map string to list of string`, func(s *testcase.Spec) {
			httpspec.LetBody(s, func(t *testcase.T) interface{} {
				return map[string][]string{"hello": {`a`, `b`, `c`}}
			})

			s.And(`form encoding is used`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
				})

				s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)

					require.Equal(t, `hello=a&hello=b&hello=c`, string(body))
				})
			})

			s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header(t).Set(`Content-Type`, `application/json`)
				})

				s.Then(`it will use json encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)
					expected := `{"hello":["a","b","c"]}` + "\n"
					actual := string(body)
					require.Equal(t, len(expected), len(actual))
					require.Equal(t, expected, actual)
				})
			})
		})

		s.Context(`as url.Values`, func(s *testcase.Spec) {
			httpspec.LetBody(s, func(t *testcase.T) interface{} {
				return url.Values{"foo": {"baz", "bar"}}
			})

			s.And(`form encoding is used`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
				})

				s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)

					require.Equal(t, `foo=baz&foo=bar`, string(body))
				})
			})

			s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header(t).Set(`Content-Type`, `application/json`)
				})

				s.Then(`it will use json encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)
					expected := `{"foo":["baz","bar"]}` + "\n"
					actual := string(body)
					require.Equal(t, len(expected), len(actual))
					require.Equal(t, expected, actual)
				})
			})
		})
	})
}
