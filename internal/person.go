package internal

type PersonOption interface {
	configure(*PersonConfig)
}

func ToPersonConfig(opts ...PersonOption) PersonConfig {
	var c PersonConfig
	for _, opt := range opts {
		opt.configure(&c)
	}
	return c
}

type PersonConfig struct {
	SexType SexType
}

type SexType int

func (st SexType) configure(c *PersonConfig) {
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
