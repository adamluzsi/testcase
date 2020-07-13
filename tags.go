package testcase

import (
	"os"
	"strings"
	"sync"
)

type tagSettings struct {
	Include map[string]struct{}
	Exclude map[string]struct{}
}

const (
	envKeyTagIncludeList = `TESTCASE_TAG_INCLUDE`
	envKeyTagExcludeList = `TESTCASE_TAG_EXCLUDE`
)

func getTagSettings() tagSettings {
	var settings = tagSettings{
		Include: map[string]struct{}{},
		Exclude: map[string]struct{}{},
	}

	if rawList, ok := os.LookupEnv(envKeyTagIncludeList); ok {
		for _, rawTag := range strings.Split(rawList, `,`) {
			tag := strings.TrimSpace(rawTag)
			settings.Include[tag] = struct{}{}
		}
	}

	if rawList, ok := os.LookupEnv(envKeyTagExcludeList); ok {
		for _, rawTag := range strings.Split(rawList, `,`) {
			tag := strings.TrimSpace(rawTag)
			settings.Exclude[tag] = struct{}{}
		}
	}

	return settings
}

var (
	tagSettingsSetup sync.Once
	tagSettingsCache tagSettings
)

func getCachedTagSettings() tagSettings {
	tagSettingsSetup.Do(func() {
		tagSettingsCache = getTagSettings()
	})

	return tagSettingsCache
}
