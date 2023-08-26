package testcase

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/adamluzsi/testcase/assert"
)

func TestSpec_Tag_withEnvVariable(t *testing.T) {
	defer resetTagEnvVariables()()

	t.Run(fmt.Sprintf(`when tags used to select certain tests`), func(t *testing.T) {
		const envKey = `TESTCASE_TAG_INCLUDE`

		includeCases := func(t *testing.T) {
			t.Run(`and the spec do not have tags that was selected`, func(t *testing.T) {
				assertTestRan(t, func(s *Spec) {}, false)
			})

			t.Run(`and the spec have at least one tag, but it is different`, func(t *testing.T) {
				assertTestRan(t, func(s *Spec) { s.Tag(`b-tag`, `c-tag`) }, false)
			})

			t.Run(`and spec have the given tag`, func(t *testing.T) {
				assertTestRan(t, func(s *Spec) { s.Tag(`the-tag`, `b-tag`) }, true)
			})
		}

		t.Run(`and only a single value is present in the tag list`, func(t *testing.T) {
			defer resetTagEnvVariables()()
			os.Setenv(envKey, `the-tag`)
			includeCases(t)
		})

		t.Run(`and multiple tag is given in the include tag list separated by comma`, func(t *testing.T) {
			defer resetTagEnvVariables()()
			os.Setenv(envKey, `the-tag,z-tag`)
			includeCases(t)
		})

		t.Run(`and multiple tag is given in the include tag list separated by comma and spacing`, func(t *testing.T) {
			defer resetTagEnvVariables()()
			os.Setenv(envKey, `the-tag, z-tag`)
			includeCases(t)
		})
	})

	t.Run(fmt.Sprintf(`when tags used to exclude certain tests`), func(t *testing.T) {
		const envKey = `TESTCASE_TAG_EXCLUDE`

		includeCases := func(t *testing.T) {
			t.Run(`and the spec do not have tags`, func(t *testing.T) {
				assertTestRan(t, func(s *Spec) {}, true)
			})

			t.Run(`and the spec have at least one tag but id does not match the excluded tag`, func(t *testing.T) {
				assertTestRan(t, func(s *Spec) { s.Tag(`b-tag`, `c-tag`) }, true)
			})

			t.Run(`and spec have the given tag that match the exclude list`, func(t *testing.T) {
				assertTestRan(t, func(s *Spec) { s.Tag(`the-tag`, `b-tag`) }, false)
			})
		}

		t.Run(`and only a single value is present in the tag list`, func(t *testing.T) {
			defer resetTagEnvVariables()()
			os.Setenv(envKey, `the-tag`)
			includeCases(t)
		})

		t.Run(`and multiple tag is given in the include tag list separated by comma`, func(t *testing.T) {
			defer resetTagEnvVariables()()
			os.Setenv(envKey, `the-tag,z-tag`)
			includeCases(t)
		})

		t.Run(`and multiple tag is given in the include tag list separated by comma and spacing`, func(t *testing.T) {
			defer resetTagEnvVariables()()
			os.Setenv(envKey, `the-tag, z-tag`)
			includeCases(t)
		})
	})
}

func resetEnv(key string) func() {
	ogValue, ok := os.LookupEnv(key)

	return func() {
		if ok {
			os.Setenv(key, ogValue)
		} else {
			os.Unsetenv(key)
		}
	}
}

func resetTagEnvVariables() func() {
	ilr := resetEnv(envKeyTagIncludeList)
	elr := resetEnv(envKeyTagExcludeList)
	return func() {
		ilr()
		elr()
	}
}

func resetTagCache() {
	tagSettingsCache = tagSettings{}
	tagSettingsSetup = sync.Once{}
}

func assertTestRan(t *testing.T, setup func(s *Spec), expected bool) {
	resetTagCache()
	defer resetTagCache()

	var actually bool
	var arrange = func(s *Spec) {
		s.Test(``, func(t *T) { actually = true })
	}

	var modifier string
	if !expected {
		modifier = `not `
	}

	t.Run(fmt.Sprintf(`then it is expected to %srun`, modifier), func(t *testing.T) {
		var s *Spec
		t.Run(``, func(t *testing.T) {
			s = NewSpec(t)
			setup(s)
			arrange(s)
		})
		assert.Must(t).Equal(expected, actually)
	})

	t.Run(`and when tags applied in a sub spec`, func(t *testing.T) {
		t.Run(fmt.Sprintf(`then it is expected to %srun in sub spec as well`, modifier), func(t *testing.T) {
			t.Run(``, func(t *testing.T) {
				parent := NewSpec(t)
				parent.Context(``, func(s *Spec) {
					setup(s)
					arrange(s)
				})
			})

			assert.Must(t).Equal(expected, actually)
		})
	})

	t.Run(`and when tags applied in parent spec`, func(t *testing.T) {
		t.Run(``, func(t *testing.T) {
			parent := NewSpec(t)
			setup(parent)
			parent.Context(``, func(s *Spec) { arrange(s) })
		})

		assert.Must(t).Equal(expected, actually,
			assert.Message(fmt.Sprintf(`then it is expected to %srun in sub spec as well`, modifier)))
	})
}
