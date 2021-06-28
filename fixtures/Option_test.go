package fixtures_test

import (
	"testing"

	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
)

func TestSkipTag(t *testing.T) {
	type Entity struct {
		First  string `other:"value" json:"other"`
		Second string `custom-tag:"foo,bar" json:"custom_one"`
		Third  string `custom-tag:"bar,baz" json:"custom_two"`
	}

	t.Run(`.New`, func(t *testing.T) {
		t.Run(`when tag specified`, func(t *testing.T) {
			ent := fixtures.New(Entity{}, fixtures.SkipByTag(`custom-tag`)).(*Entity)
			require.NotEmpty(t, ent.First)
			require.Empty(t, ent.Second)
			require.Empty(t, ent.Third)
		})

		t.Run(`when value of a tag is specified`, func(t *testing.T) {
			ent := fixtures.New(Entity{}, fixtures.SkipByTag(`custom-tag`, "baz")).(*Entity)
			require.NotEmpty(t, ent.First)
			require.NotEmpty(t, ent.Second)
			require.Empty(t, ent.Third)
		})

		t.Run(`when multiple value of a given tag is specified`, func(t *testing.T) {
			ent := fixtures.New(Entity{}, fixtures.SkipByTag(`custom-tag`, "foo", "baz")).(*Entity)
			require.NotEmpty(t, ent.First)
			require.Empty(t, ent.Second)
			require.Empty(t, ent.Third)
		})
	})

	t.Run(`.Create`, func(t *testing.T) {
		subject := func(tb testing.TB, options ...fixtures.Option) Entity {
			ff := &fixtures.Factory{Options: options}
			return ff.Create(Entity{}).(Entity)
		}

		t.Run(`when tag specified`, func(t *testing.T) {
			ent := subject(t, fixtures.SkipByTag(`custom-tag`))
			require.NotEmpty(t, ent.First)
			require.Empty(t, ent.Second)
			require.Empty(t, ent.Third)
		})

		t.Run(`when value of a tag is specified`, func(t *testing.T) {
			ent := subject(t, fixtures.SkipByTag(`custom-tag`, "baz"))
			require.NotEmpty(t, ent.First)
			require.NotEmpty(t, ent.Second)
			require.Empty(t, ent.Third)
		})

		t.Run(`when multiple value of a given tag is specified`, func(t *testing.T) {
			ent := subject(t, fixtures.SkipByTag(`custom-tag`, "foo", "baz"))
			require.NotEmpty(t, ent.First)
			require.Empty(t, ent.Second)
			require.Empty(t, ent.Third)
		})
	})
}
