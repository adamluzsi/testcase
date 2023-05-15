package internal

type ContactOption interface {
	configure(*ContactConfig)
}

func ToContactConfig(opts ...ContactOption) ContactConfig {
	var c ContactConfig
	for _, opt := range opts {
		opt.configure(&c)
	}
	return c
}

type ContactConfig struct {
	SexType SexType
}

type SexType int

func (st SexType) configure(c *ContactConfig) {
	if c.SexType == 0 {
		c.SexType = st
		return
	}
	if c.SexType == st {
		return
	}
	if c.SexType != st {
		c.SexType = SexTypeAny
		return
	}
}

const (
	_ SexType = iota
	SexTypeMale
	SexTypeFemale
	SexTypeAny
)
