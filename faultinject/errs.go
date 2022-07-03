package faultinject

type errT string

func (err errT) Error() string { return string(err) }

const DefaultErr errT = "fault injected"
