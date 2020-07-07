package testcase

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
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

	var expectRanTo = func(s *Spec, expectedRan bool) {
		var actuallyRan bool
		s.Test(``, func(t *T) { actuallyRan = true })
		require.Equal(t, expectedRan, actuallyRan)
	}

	var modifier string
	if !expected {
		modifier = `not `
	}

	t.Run(fmt.Sprintf(`then it is expected to %srun`, modifier), func(t *testing.T) {
		s := NewSpec(t)
		setup(s)

		expectRanTo(s, expected)
	})

	t.Run(`and when tags applied in a sub context`, func(t *testing.T) {
		parent := NewSpec(t)
		var sub *Spec
		parent.Context(``, func(s *Spec) { sub = s })

		t.Run(fmt.Sprintf(`then it is expected to %srun in sub context as well`, modifier), func(t *testing.T) {
			setup(sub)

			expectRanTo(sub, expected)
		})
	})

	t.Run(`and when tags applied in parent context`, func(t *testing.T) {
		parent := NewSpec(t)
		setup(parent)
		var sub *Spec
		parent.Context(``, func(s *Spec) { sub = s })

		t.Run(fmt.Sprintf(`then it is expected to %srun in sub context as well`, modifier), func(t *testing.T) {
			expectRanTo(sub, expected)
		})
	})
}
