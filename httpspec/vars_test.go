package httpspec_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/httpspec"
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
	httpspec.Handler.Let(s, func(t *testcase.T) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx = r.Context()
			method = r.Method
			path = r.URL.Path
			query = r.URL.Query()
			header = r.Header
			bs, err := ioutil.ReadAll(r.Body)
			assert.Must(t).Nil(err)
			assert.Must(t).Nil(r.Body.Close())
			body = bs
			w.Header().Set(`Hello`, `World`)
			w.WriteHeader(http.StatusTeapot)
			_, err = fmt.Fprint(w, `Hello, World!`)
			assert.Must(t).Nil(err)
		})
	})

	s.Describe(`httpspec.ServeHTTP`, func(s *testcase.Spec) {
		s.Then(`it should return a response recorder with the API response`, func(t *testcase.T) {
			rr := httpspec.ServeHTTP(t)
			t.Must.Equal(http.StatusTeapot, rr.Code)
			t.Must.Equal(`World`, rr.Header().Get(`Hello`))
			t.Must.Equal(`Hello, World!`, rr.Body.String())
		})
	})

	s.When(`context defined`, func(s *testcase.Spec) {
		var expected = context.WithValue(context.Background(), `key`, `value`)
		httpspec.Context.Let(s, func(t *testcase.T) context.Context { return expected })

		s.And(`using context key-value is added with testcase.T#Let + httpspec.ContextVarName`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				httpspec.Context.Set(t, context.WithValue(httpspec.Context.Get(t), `foo`, `bar`))
			})

			s.Then(`in this scope the key-values of the context will be updated`, func(t *testcase.T) {
				httpspec.ServeHTTP(t)

				t.Must.Equal(`bar`, ctx.Value(`foo`).(string))
			})
		})

		s.Then(`the context will be passed for the request`, func(t *testcase.T) {
			t.Log(`this can be used to create API specs where value in context is part of the http.handler prerequisite`)
			httpspec.ServeHTTP(t)
			t.Must.Equal(expected, ctx)
		})
	})

	s.When(`query populated during Spec#Before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			httpspec.Query.Get(t).Set(`hello`, `world`)
			httpspec.Query.Get(t).Add(`l`, `a`)
			httpspec.Query.Get(t).Add(`l`, `b`)
			httpspec.Query.Get(t).Add(`l`, `c`)
		})

		s.Then(`it will pass the query to the request`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			t.Must.ContainExactly([]string{`a`, `b`, `c`}, query[`l`])
			t.Must.Equal(`world`, query.Get(`hello`))
		})
	})

	s.When(`header populated during Spec#Before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			httpspec.Header.Get(t).Set(`Hello`, `world`)
			httpspec.Header.Get(t).Add(`L`, `a`)
			httpspec.Header.Get(t).Add(`L`, `b`)
			httpspec.Header.Get(t).Add(`L`, `c`)
		})

		s.Then(`it will HandlerLet the headers for the request`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			t.Log(header)
			t.Must.ContainExactly([]string{`a`, `b`, `c`}, header[`L`])
			t.Must.Equal(`world`, header.Get(`Hello`))
		})
	})

	s.When(`PathGet is not defined`, func(s *testcase.Spec) {
		s.Then(`it will use / as default`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			t.Must.Equal(`/`, path)
		})
	})

	s.When(`PathGet is defined with PathLet`, func(s *testcase.Spec) {
		httpspec.Path.Let(s, func(t *testcase.T) string { return `/hello/world` })

		s.Then(`it will call request with the given PathGet`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			t.Must.Equal(`/hello/world`, path)
		})
	})

	s.When(`PathGet is defined with PathLetValue`, func(s *testcase.Spec) {
		httpspec.Path.LetValue(s, `/foo/baz`)

		s.Then(`it will call request with the given PathGet`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			t.Must.Equal(`/foo/baz`, path)
		})
	})

	s.When(`MethodGet is not defined`, func(s *testcase.Spec) {
		s.Then(`it will use HTTP GET as default`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			t.Must.Equal(http.MethodGet, method)
		})
	})

	s.When(`MethodGet is defined with MethodLet`, func(s *testcase.Spec) {
		httpspec.Method.Let(s, func(t *testcase.T) string { return http.MethodPost })

		s.Then(`it will use the http MethodGet for the request`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			t.Must.Equal(http.MethodPost, method)
		})
	})

	s.When(`MethodGet is defined with MethodLetValue`, func(s *testcase.Spec) {
		httpspec.Method.LetValue(s, http.MethodPut)

		s.Then(`it will use the http MethodGet for the request`, func(t *testcase.T) {
			httpspec.ServeHTTP(t)
			t.Must.Equal(http.MethodPut, method)
		})
	})

	s.When(`body defined`, func(s *testcase.Spec) {
		const expected = `Hello, World!`

		s.Context(`as io.Reader`, func(s *testcase.Spec) {
			httpspec.Body.Let(s, func(t *testcase.T) interface{} {
				return strings.NewReader(`Hello, World!`)
			})

			s.Then(`value is passed as is, without any further action`, func(t *testcase.T) {
				httpspec.ServeHTTP(t)
				actual := string(body)
				t.Must.Equal(len(expected), len(actual))
				t.Must.Equal(expected, actual)
			})

			s.And(`if debugging enabled`, func(s *testcase.Spec) {
				httpspec.Debug(s)

				s.Then(`it will pass the io reader content`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)
					actual := string(body)
					t.Must.Equal(len(expected), len(actual))
					t.Must.Equal(expected, actual)
				})
			})
		})

		s.Context(`as struct`, func(s *testcase.Spec) {
			s.And(`it has tags for form and json to define the keys`, func(s *testcase.Spec) {
				httpspec.Body.Let(s, func(t *testcase.T) interface{} {
					return struct {
						Hello string `json:"hello_json_key" form:"hello_form_key"`
					}{Hello: `world`}
				})

				s.And(`form encoding is used`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header.Get(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
					})

					s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)

						t.Must.Equal(`hello_form_key=world`, string(body))
					})
				})

				s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header.Get(t).Set(`Content-Type`, `application/json`)
					})

					s.Then(`it will use json encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)
						expected := `{"hello_json_key":"world"}` + "\n"
						actual := string(body)
						t.Must.Equal(len(expected), len(actual))
						t.Must.Equal(expected, actual)
					})
				})
			})

			s.And(`it has no tags`, func(s *testcase.Spec) {
				httpspec.Body.Let(s, func(t *testcase.T) interface{} {
					return struct{ TheKey string }{TheKey: `TheValue`}
				})

				s.And(`form encoding is used`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header.Get(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
					})

					s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)

						t.Must.Equal(`TheKey=TheValue`, string(body))
					})
				})

				s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header.Get(t).Set(`Content-Type`, `application/json`)
					})

					s.Then(`it will use json encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)
						expected := `{"TheKey":"TheValue"}` + "\n"
						actual := string(body)
						t.Must.Equal(len(expected), len(actual))
						t.Must.Equal(expected, actual)
					})
				})
			})
		})

		s.Context(`as pointer`, func(s *testcase.Spec) {
			s.Context(`to struct`, func(s *testcase.Spec) {
				httpspec.Body.Let(s, func(t *testcase.T) interface{} {
					return &struct {
						Hello string `json:"hello_json" form:"hello_form"`
					}{Hello: `world`}
				})

				s.And(`form encoding is used`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header.Get(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
					})

					s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)

						t.Must.Equal(`hello_form=world`, string(body))
					})
				})

				s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						httpspec.Header.Get(t).Set(`Content-Type`, `application/json`)
					})

					s.Then(`it will use json encoding`, func(t *testcase.T) {
						httpspec.ServeHTTP(t)
						expected := `{"hello_json":"world"}` + "\n"
						actual := string(body)
						t.Must.Equal(len(expected), len(actual))
						t.Must.Equal(expected, actual)
					})
				})
			})
		})

		s.Context(`as map string to string`, func(s *testcase.Spec) {
			httpspec.Body.Let(s, func(t *testcase.T) interface{} {
				return map[string]string{"hello": "world"}
			})

			s.And(`form encoding is used`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header.Get(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
				})

				s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)

					t.Must.Equal(`hello=world`, string(body))
				})
			})

			s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header.Get(t).Set(`Content-Type`, `application/json`)
				})

				s.Then(`it will use json encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)
					expected := `{"hello":"world"}` + "\n"
					actual := string(body)
					t.Must.Equal(len(expected), len(actual))
					t.Must.Equal(expected, actual)
				})
			})
		})

		s.Context(`as map string to list of string`, func(s *testcase.Spec) {
			httpspec.Body.Let(s, func(t *testcase.T) interface{} {
				return map[string][]string{"hello": {`a`, `b`, `c`}}
			})

			s.And(`form encoding is used`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header.Get(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
				})

				s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)

					t.Must.Equal(`hello=a&hello=b&hello=c`, string(body))
				})
			})

			s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header.Get(t).Set(`Content-Type`, `application/json`)
				})

				s.Then(`it will use json encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)
					expected := `{"hello":["a","b","c"]}` + "\n"
					actual := string(body)
					t.Must.Equal(len(expected), len(actual))
					t.Must.Equal(expected, actual)
				})
			})
		})

		s.Context(`as url.Values`, func(s *testcase.Spec) {
			httpspec.Body.Let(s, func(t *testcase.T) interface{} {
				return url.Values{"foo": {"baz", "bar"}}
			})

			s.And(`form encoding is used`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header.Get(t).Set(`Content-Type`, `application/x-www-form-urlencoded`)
				})

				s.Then(`it will use over simplified basic form url encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)

					t.Must.Equal(`foo=baz&foo=bar`, string(body))
				})
			})

			s.And(`json encoding is used for the request`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					httpspec.Header.Get(t).Set(`Content-Type`, `application/json`)
				})

				s.Then(`it will use json encoding`, func(t *testcase.T) {
					httpspec.ServeHTTP(t)
					expected := `{"foo":["baz","bar"]}` + "\n"
					actual := string(body)
					t.Must.Equal(len(expected), len(actual))
					t.Must.Equal(expected, actual)
				})
			})
		})
	})
}
