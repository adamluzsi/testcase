package fixtures_test

import (
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSkipTag(t *testing.T) {
	type Entity struct {
		First  string `other:"value" json:"other"`
		Second string `custom-tag:"foo,bar" json:"custom_one"`
		Third  string `custom-tag:"bar,baz" json:"custom_two"`
	}

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
}
