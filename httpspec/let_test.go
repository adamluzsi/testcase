package httpspec_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
	"github.com/adamluzsi/testcase/random"
)

func TestLetResponseRecorder(t *testing.T) {
	s := testcase.NewSpec(t)
	rr := httpspec.LetResponseRecorder(s)
	s.Test("", func(t *testcase.T) {
		t.Must.Empty(rr.Get(t).Body.String())
		_, err := rr.Get(t).WriteString("hello")
		t.Must.NoError(err)
		t.Must.Contain(rr.Get(t).Body.String(), "hello")
	})
}

func TestLetRequest(t *testing.T) {
	s := testcase.NewSpec(t)

	s.When("RequestVar is empty", func(s *testcase.Spec) {
		request := httpspec.LetRequest(s, httpspec.RequestVar{})
		s.Then("default values used", func(t *testcase.T) {
			r := request.Get(t)
			t.Must.NotEmpty(r.Host)
			t.Must.Equal(http.MethodGet, r.Method)
			t.Must.Equal("/", r.URL.Path)
			t.Must.Empty(r.URL.Query())
			t.Must.Empty(r.Header)
			t.Must.Read("", r.Body)
		})
	})

	s.When("RequestVar is populated with values", func(s *testcase.Spec) {
		type BodyDTO struct {
			V1 string `json:"v1"`
			V2 int    `json:"v2"`
		}
		rv := httpspec.RequestVar{
			Context: testcase.Let(s, func(t *testcase.T) context.Context {
				return context.WithValue(context.Background(), "foo", "bar")
			}),
			Scheme: testcase.Let(s, func(t *testcase.T) string {
				return "postgres"
			}),
			Host: testcase.Let(s, func(t *testcase.T) string {
				return fmt.Sprintf("www.%s.com", t.Random.StringNC(5, random.CharsetAlpha()))
			}),
			Method: testcase.Let(s, func(t *testcase.T) string {
				return t.Random.ElementFromSlice([]string{
					http.MethodGet,
					http.MethodPost,
					http.MethodPut,
					http.MethodDelete,
				}).(string)
			}),
			Path: testcase.Let(s, func(t *testcase.T) string {
				cs := random.CharsetAlpha()
				return "/" + path.Join(t.Random.StringNC(3, cs), t.Random.StringNC(5, cs))
			}),
			Query: testcase.Let(s, func(t *testcase.T) url.Values {
				return url.Values{t.Random.StringNC(5, random.CharsetAlpha()): []string{t.Random.StringN(5)}}
			}),
			Header: testcase.Let(s, func(t *testcase.T) http.Header {
				h := http.Header{}
				charset := random.CharsetAlpha()
				h.Set(t.Random.StringNC(5, charset), t.Random.StringNC(5, charset))
				h.Set("Content-Type", "application/json")
				return h
			}),
			Body: testcase.Let(s, func(t *testcase.T) any {
				return BodyDTO{
					V1: t.Random.String(),
					V2: t.Random.Int(),
				}
			}),
		}
		request := httpspec.LetRequest(s, rv)

		s.Test("injected variables used", func(t *testcase.T) {
			r := request.Get(t)
			t.Must.Equal(rv.Header.Get(t), r.Header)
			t.Must.Equal(rv.Path.Get(t), r.URL.Path)
			t.Must.Equal(rv.Query.Get(t), r.URL.Query())
			t.Must.Equal(rv.Scheme.Get(t), r.URL.Scheme)
			t.Must.Equal(rv.Method.Get(t), r.Method)
			t.Must.Equal(rv.Context.Get(t), r.Context())
			var body BodyDTO
			t.Must.NoError(json.Unmarshal(t.Must.ReadAll(r.Body), &body))
			t.Must.Equal(rv.Body.Get(t), body)
		})
	})
}
