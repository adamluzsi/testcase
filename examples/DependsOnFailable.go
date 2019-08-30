package examples

type DependsOnFailable struct {
	Failable Failable
}

func (d *DependsOnFailable) Run() error {
	return d.Failable.Do()
}

type Failable interface {
	Do() error
}
