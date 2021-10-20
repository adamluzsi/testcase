package fixtures_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/fixtures"
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
			assert.Must(t).NotEqual(len(ent.First), 0)
			assert.Must(t).True(len(ent.Second) == 0)
			assert.Must(t).True(len(ent.Third) == 0)
		})

		t.Run(`when value of a tag is specified`, func(t *testing.T) {
			ent := fixtures.New(Entity{}, fixtures.SkipByTag(`custom-tag`, "baz")).(*Entity)
			assert.Must(t).True(0 < len(ent.First))
			assert.Must(t).True(0 < len(ent.Second))
			assert.Must(t).True(len(ent.Third) == 0)
		})

		t.Run(`when multiple value of a given tag is specified`, func(t *testing.T) {
			ent := fixtures.New(Entity{}, fixtures.SkipByTag(`custom-tag`, "foo", "baz")).(*Entity)
			assert.Must(t).True(0 < len(ent.First))
			assert.Must(t).True(len(ent.Second) == 0)
			assert.Must(t).True(len(ent.Third) == 0)
		})
	})

	t.Run(`.Fixture`, func(t *testing.T) {
		subject := func(tb testing.TB, options ...fixtures.Option) Entity {
			ff := &fixtures.Factory{Options: options}
			ctx := context.Background()
			return ff.Fixture(Entity{}, ctx).(Entity)
		}

		t.Run(`when tag specified`, func(t *testing.T) {
			ent := subject(t, fixtures.SkipByTag(`custom-tag`))
			assert.Must(t).True(0 < len(ent.First))
			assert.Must(t).True(len(ent.Second) == 0)
			assert.Must(t).True(len(ent.Third) == 0)
		})

		t.Run(`when value of a tag is specified`, func(t *testing.T) {
			ent := subject(t, fixtures.SkipByTag(`custom-tag`, "baz"))
			assert.Must(t).True(0 < len(ent.First))
			assert.Must(t).True(0 < len(ent.Second))
			assert.Must(t).True(len(ent.Third) == 0)
		})

		t.Run(`when multiple value of a given tag is specified`, func(t *testing.T) {
			ent := subject(t, fixtures.SkipByTag(`custom-tag`, "foo", "baz"))
			assert.Must(t).True(0 < len(ent.First))
			assert.Must(t).True(len(ent.Second) == 0)
			assert.Must(t).True(len(ent.Third) == 0)
		})
	})
}
