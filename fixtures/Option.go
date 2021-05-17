package fixtures

import (
	"reflect"
	"regexp"
)

type Options []Option

func (os Options) Testing() {}

type Option interface{ setup(c *Config) }

type optionFunc func(c *Config)

func (fn optionFunc) setup(c *Config) { fn(c) }

func NewConfig(opts ...Option) *Config {
	var c Config
	for _, opt := range opts {
		opt.setup(&c)
	}
	return &c
}

type Config struct {
	skipByTags map[string][]string // tag -> values
}

func (c *Config) GetSkipTags() map[string][]string {
	if c.skipByTags == nil {
		c.skipByTags = make(map[string][]string)
	}
	return c.skipByTags
}

var structFieldTagSeparator = regexp.MustCompile(`,|;`)

func (c *Config) CanPopulateStructField(sf reflect.StructField) bool {
	for tagName, values := range c.GetSkipTags() {
		tag, ok := sf.Tag.Lookup(tagName)
		if !ok {
			continue
		}

		if len(values) == 0 {
			return false
		}

		tagValueIndex := make(map[string]struct{})

		for _, v := range structFieldTagSeparator.Split(tag, -1) {
			tagValueIndex[v] = struct{}{}
		}

		for _, value := range values {
			if _, ok := tagValueIndex[value]; ok {
				return false
			}
		}
	}

	return true
}

// SkipByTag is an Option to skip a certain tag during the New function value population.
// If value is not provided, all matching tag will be skipped.
// If value or multiple value is provided, then matching tag only skipped if it matches the values.
func SkipByTag(tag string, values ...string) Option {
	return optionFunc(func(c *Config) {
		c.GetSkipTags()[tag] = append(c.GetSkipTags()[tag], values...)
	})
}
